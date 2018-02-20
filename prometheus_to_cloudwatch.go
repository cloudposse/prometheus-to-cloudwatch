package main

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"log"
)

const (
	batchSize      = 20
	cwHighResLabel = "__cw_high_res"
	cwUnitLabel    = "__cw_unit"
)

// Config defines configuration options
type Config struct {
	// Required. The CloudWatch namespace under which metrics should be published
	CloudWatchNamespace string

	// Required. The AWS Region to use
	CloudWatchRegion string

	// The frequency with which metrics should be published to Cloudwatch. Default: 15s
	CloudWatchPublishInterval time.Duration

	// Timeout for sending metrics to Cloudwatch. Default: 3s
	CloudWatchPublishTimeout time.Duration
}

// Bridge pushes metrics to AWS CloudWatch
type Bridge struct {
	cloudWatchPublishInterval time.Duration
	cloudWatchNamespace       string
	gatherer                  prometheus.Gatherer
	cw                        *cloudwatch.CloudWatch
}

// NewBridge initializes and returns a pointer to a Bridge using the
// supplied configuration, or an error if there is a problem with
// the configuration
func NewBridge(c *Config) (*Bridge, error) {
	b := &Bridge{}

	if c.CloudWatchNamespace == "" {
		return nil, errors.New("CloudWatchNamespace must not be empty")
	}
	b.cloudWatchNamespace = c.CloudWatchNamespace

	if c.CloudWatchPublishInterval > 0 {
		b.cloudWatchPublishInterval = c.CloudWatchPublishInterval
	} else {
		b.cloudWatchPublishInterval = 15 * time.Second
	}

	var client = http.DefaultClient

	if c.CloudWatchPublishTimeout > 0 {
		client.Timeout = c.CloudWatchPublishTimeout
	} else {
		client.Timeout = 3 * time.Second
	}

	b.gatherer = prometheus.DefaultGatherer

	// Use default credential provider, which supports the standard
	// AWS_* environment variables, and the shared credential file under ~/.aws
	sess, err := session.NewSession(aws.NewConfig().WithHTTPClient(client).WithRegion(c.CloudWatchRegion))
	if err != nil {
		return nil, err
	}

	b.cw = cloudwatch.New(sess)
	return b, nil
}

// Run starts a loop that will push metrics to Cloudwatch at the
// configured interval. Accepts a context.Context to support cancellation
func (b *Bridge) Run(ctx context.Context) {
	ticker := time.NewTicker(b.cloudWatchPublishInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mfs, err := b.gatherer.Gather()
			if err != nil {
				log.Println("prometheus-to-cloudwatch: error gathering metrics from Prometheus:", err)
			}
			err = b.publishMetrics(mfs)
			if err != nil {
				log.Println("prometheus-to-cloudwatch: error publishing to Cloudwatch:", err)
			}
		case <-ctx.Done():
			log.Println("prometheus-to-cloudwatch: stopping")
			return
		}
	}
}

// NOTE: The CloudWatch API has the following limitations:
//		- Max 40kb request size
//		- Single namespace per request
//		- Max 10 dimensions per metric
func (b *Bridge) publishMetrics(mfs []*dto.MetricFamily) error {
	vec, err := expfmt.ExtractSamples(&expfmt.DecodeOptions{Timestamp: model.Now()}, mfs...)

	if err != nil {
		return err
	}

	data := make([]*cloudwatch.MetricDatum, 0, batchSize)

	for _, s := range vec {
		name := getName(s.Metric)
		data = appendDatum(data, name, s)

		// 40KB CloudWatch size limitation
		if len(data) == batchSize {
			if err := b.flush(data); err != nil {
				log.Println("prometheus-to-cloudwatch: error publishing to Cloudwatch:", err)
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
			Namespace:  &b.cloudWatchNamespace,
		}
		_, err := b.cw.PutMetricData(in)
		return err
	}
	return nil
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

func getUnit(m model.Metric) string {
	if u, ok := m[cwUnitLabel]; ok {
		return string(u)
	}
	return "None"
}
