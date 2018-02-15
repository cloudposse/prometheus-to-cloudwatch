package main

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

const (
	batchSize = 20

	cwHighResLabel = "__cw_high_res"
	cwUnitLabel    = "__cw_unit"
)

// Config defines configuration options for Bridge
type Config struct {
	// Required. The Prometheus namespace/prefix to scrape. Each Bridge only supports 1 prefix.
	// If multiple prefixes are required, multiple Bridges must be used.
	PrometheusNamespace string

	// Required. The CloudWatch namespace under which metrics should be published
	CloudWatchNamespace string

	// Required. The AWS Region to use
	CloudWatchRegion string

	// The frequency with which metrics should be published to Cloudwatch. Default: 15s
	Interval time.Duration

	// Timeout for sending metrics to Cloudwatch. Default: 1s
	Timeout time.Duration

	// Custom HTTP Client to use with the Cloudwatch API. Default: http.Client{}
	// If Config.Timeout is supplied, it will override any timeout defined on
	// the supplied http.Client
	Client *http.Client

	// Logger that messages are written to. Default: nil
	Logger Logger

	// The Gatherer to use for metrics. Default: prometheus.DefaultGatherer
	Gatherer prometheus.Gatherer

	// Only publish whitelisted metrics
	WhitelistOnly bool

	// List of metrics that should be published, causing all others to be ignored.
	// Config.WhitelistOnly must be set to true for this to take effect.
	Whitelist []string

	// List of metrics that should never be published. This setting overrides entries in Config.Whitelist
	Blacklist []string
}

// Bridge pushes metrics to AWS Cloudwatch
type Bridge struct {
	interval time.Duration
	timeout  time.Duration

	promNamespace string
	cwNamespace   string

	useWhitelist bool
	whitelist    map[string]struct{}
	blacklist    map[string]struct{}

	logger Logger
	g      prometheus.Gatherer
	cw     *cloudwatch.CloudWatch
}

// NewBridge initializes and returns a pointer to a Bridge using the
// supplied configuration, or an error if there is a problem with
// the configuration
func NewBridge(c *Config) (*Bridge, error) {
	b := &Bridge{}

	if c.PrometheusNamespace == "" {
		return nil, errors.New("PrometheusNamespace must not be empty")
	}
	b.promNamespace = c.PrometheusNamespace

	if c.CloudWatchNamespace == "" {
		return nil, errors.New("CloudWatchNamespace must not be empty")
	}
	b.cwNamespace = c.CloudWatchNamespace

	if c.Interval > 0 {
		b.interval = c.Interval
	} else {
		b.interval = 15 * time.Second
	}

	var client *http.Client
	if c.Client != nil {
		client = c.Client
	} else {
		client = &http.Client{}
	}

	if c.Timeout > 0 {
		client.Timeout = c.Timeout
	} else {
		client.Timeout = time.Second
	}

	if c.Logger != nil {
		b.logger = c.Logger
	}

	if c.Gatherer != nil {
		b.g = c.Gatherer
	} else {
		b.g = prometheus.DefaultGatherer
	}

	b.useWhitelist = c.WhitelistOnly
	b.whitelist = make(map[string]struct{}, len(c.Whitelist))
	for _, v := range c.Whitelist {
		b.whitelist[v] = struct{}{}
	}

	b.blacklist = make(map[string]struct{}, len(c.Blacklist))
	for _, v := range c.Blacklist {
		b.blacklist[v] = struct{}{}
	}

	// Use default credential provider, which I believe supports the standard
	// AWS_* environment variables, and the shared credential file under ~/.aws
	sess, err := session.NewSession(aws.NewConfig().WithHTTPClient(client).WithRegion(c.CloudWatchRegion))
	if err != nil {
		return nil, err
	}

	b.cw = cloudwatch.New(sess)
	return b, nil
}

// Logger is the minimal interface Bridge needs for logging. Note that
// log.Logger from the standard library implements this interface, and it is
// easy to implement by custom loggers, if they don't do so already anyway.
// Taken from https://github.com/prometheus/client_golang/blob/master/prometheus/graphite/bridge.go
type Logger interface {
	Println(v ...interface{})
}

// Run starts a loop that will push metrics to Cloudwatch at the
// configured interval. Run accepts a context.Context to support
// cancellation.
func (b *Bridge) Run(ctx context.Context) {
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := b.Publish(); err != nil && b.logger != nil {
				b.logger.Println("error publishing to Cloudwatch:", err)
			}
		case <-ctx.Done():
			if b.logger != nil {
				b.logger.Println("stopping Cloudwatch publisher")
			}
			return
		}
	}
}

// Publish publishes the Prometheus metrics to Cloudwatch
func (b *Bridge) Publish() error {
	mfs, err := b.g.Gather()
	if err != nil {
		return err
	}

	return b.publishMetrics(mfs)
}

// NOTE: The CloudWatch API has the following limitations:
//		- Max 40kb request size
//		- Single namespace per request
//		- Max 10 dimensions per metric
func (b *Bridge) publishMetrics(mfs []*dto.MetricFamily) error {
	vec, err := expfmt.ExtractSamples(&expfmt.DecodeOptions{
		Timestamp: model.Now(),
	}, mfs...)
	if err != nil {
		return err
	}

	data := make([]*cloudwatch.MetricDatum, 0, batchSize)
	for _, s := range vec {
		name := getName(s.Metric)
		if b.isWhitelisted(name) {
			data = appendDatum(data, name, s)
		}

		// punt on the 40KB size limitation. Will see how this works out in practice
		if len(data) == batchSize {
			if err := b.flush(data); err != nil {
				b.logger.Println("error publishing to Cloudwatch:", err)
			}
			data = make([]*cloudwatch.MetricDatum, 0, batchSize)
		}
	}

	return b.flush(data)
}

func (b *Bridge) flush(data []*cloudwatch.MetricDatum) error {
	if len(data) > 0 {
		in := &cloudwatch.PutMetricDataInput{
			MetricData: data,
			Namespace:  &b.cwNamespace,
		}
		_, err := b.cw.PutMetricData(in)
		return err
	}
	return nil
}

func (b *Bridge) isWhitelisted(name string) bool {
	if !strings.HasPrefix(name, b.promNamespace) {
		return false
	} else if _, ok := b.blacklist[name]; ok {
		return false
	}

	if b.useWhitelist {
		if name == "" {
			return false
		}
		_, ok := b.whitelist[name]
		return ok
	}
	return true
}

func appendDatum(data []*cloudwatch.MetricDatum, name string, s *model.Sample) []*cloudwatch.MetricDatum {
	d := &cloudwatch.MetricDatum{}
	d.SetMetricName(name).
		SetValue(float64(s.Value)).
		SetTimestamp(s.Timestamp.Time()).
		SetDimensions(getDimensions(s.Metric)).
		SetStorageResolution(getResolution(s.Metric)).
		SetUnit(getUnit(s.Metric))
	return append(data, d)
}

func getName(m model.Metric) string {
	if n, ok := m[model.MetricNameLabel]; ok {
		return string(n)
	}
	return ""
}

// getDimensions returns up to 10 dimensions for the provided metric - one for each label (except the __name__ label)
//
// If a metric has more than 10 labels, it attempts to behave deterministically by sorting the labels lexicographically,
// and returning the first 10 labels as dimensions
func getDimensions(m model.Metric) []*cloudwatch.Dimension {
	if len(m) == 0 {
		return make([]*cloudwatch.Dimension, 0)
	} else if _, ok := m[model.MetricNameLabel]; len(m) == 1 && ok {
		return make([]*cloudwatch.Dimension, 0)
	}
	names := make([]string, 0, len(m))
	for k := range m {
		if !(k == model.MetricNameLabel || k == cwHighResLabel || k == cwUnitLabel) {
			names = append(names, string(k))
		}
	}

	sort.Strings(names)
	if len(names) > 10 {
		names = names[:10]
	}
	dims := make([]*cloudwatch.Dimension, 0, len(names))
	for _, k := range names {
		dims = append(dims, new(cloudwatch.Dimension).SetName(k).SetValue(string(m[model.LabelName(k)])))
	}
	return dims
}

// Returns 1 if the metric contains a __cw_high_res label, otherwise it return 60
func getResolution(m model.Metric) int64 {
	if _, ok := m[cwHighResLabel]; ok {
		return 1
	}
	return 60
}

// TODO: can we infer the proper unit based on the metric name?
func getUnit(m model.Metric) string {
	if u, ok := m[cwUnitLabel]; ok {
		return string(u)
	}
	return "None"
}
