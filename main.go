package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gobwas/glob"
)

const (
	DEFAULT_LISTEN_ADDRESS = ":9698"
	DEFAULT_METRICS_PATH = "/metrics"
)

var defaultForceHighRes, _ = strconv.ParseBool(os.Getenv("FORCE_HIGH_RES"))

var (
	awsAccessKeyId              = flag.String("aws_access_key_id", os.Getenv("AWS_ACCESS_KEY_ID"), "AWS access key Id with permissions to publish CloudWatch metrics")
	awsSecretAccessKey          = flag.String("aws_secret_access_key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "AWS secret access key with permissions to publish CloudWatch metrics")
	awsSessionToken             = flag.String("aws_session_token", os.Getenv("AWS_SESSION_TOKEN"), "AWS session token with permissions to publish CloudWatch metrics")
	cloudWatchNamespace         = flag.String("cloudwatch_namespace", os.Getenv("CLOUDWATCH_NAMESPACE"), "CloudWatch Namespace")
	cloudWatchRegion            = flag.String("cloudwatch_region", os.Getenv("CLOUDWATCH_REGION"), "CloudWatch Region")
	cloudWatchPublishTimeout    = flag.String("cloudwatch_publish_timeout", os.Getenv("CLOUDWATCH_PUBLISH_TIMEOUT"), "CloudWatch publish timeout in seconds")
	prometheusScrapeInterval    = flag.String("prometheus_scrape_interval", os.Getenv("PROMETHEUS_SCRAPE_INTERVAL"), "Prometheus scrape interval in seconds")
	prometheusScrapeUrl         = flag.String("prometheus_scrape_url", os.Getenv("PROMETHEUS_SCRAPE_URL"), "Prometheus scrape URL")
	certPath                    = flag.String("cert_path", os.Getenv("CERT_PATH"), "Path to SSL Certificate file (when using SSL for `prometheus_scrape_url`)")
	keyPath                     = flag.String("key_path", os.Getenv("KEY_PATH"), "Path to Key file (when using SSL for `prometheus_scrape_url`)")
	skipServerCertCheck         = flag.String("accept_invalid_cert", os.Getenv("ACCEPT_INVALID_CERT"), "Accept any certificate during TLS handshake. Insecure, use only for testing")
	additionalDimension         = flag.String("additional_dimension", os.Getenv("ADDITIONAL_DIMENSION"), "Additional dimension specified by NAME=VALUE")
	replaceDimensions           = flag.String("replace_dimensions", os.Getenv("REPLACE_DIMENSIONS"), "replace dimensions specified by NAME=VALUE,...")
	includeMetrics              = flag.String("include_metrics", os.Getenv("INCLUDE_METRICS"), "Only publish the specified metrics (comma-separated list of glob patterns, e.g. 'up,http_*')")
	excludeMetrics              = flag.String("exclude_metrics", os.Getenv("EXCLUDE_METRICS"), "Never publish the specified metrics (comma-separated list of glob patterns, e.g. 'tomcat_*')")
	includeDimensionsForMetrics = flag.String("include_dimensions_for_metrics", os.Getenv("INCLUDE_DIMENSIONS_FOR_METRICS"), "Only publish the specified dimensions for metrics (semi-colon-separated key values of comma-separated dimensions of METRIC=dim1,dim2;, e.g. 'flink_jobmanager=job_id')")
	excludeDimensionsForMetrics = flag.String("exclude_dimensions_for_metrics", os.Getenv("EXCLUDE_DIMENSIONS_FOR_METRICS"), "Never publish the specified dimensions for metrics (semi-colon-separated key values of comma-separated dimensions of METRIC=dim1,dim2;, e.g. 'flink_jobmanager=job,host;zk_up=host,pod;')")
	forceHighRes                = flag.Bool("force_high_res", defaultForceHighRes, "Publish all metrics with high resolution, even when original metrics don't have the label "+cwHighResLabel)
	listenAddress               = flag.String("listen_address", os.Getenv("LISTEN_ADDRESS"), fmt.Sprintf("Address to expose metrics (default: %s)", DEFAULT_LISTEN_ADDRESS))
	metricsPath                 = flag.String("metrics_path", os.Getenv("METRICS_PATH"), fmt.Sprintf("Path under which to expose metrics (default: %s)", DEFAULT_METRICS_PATH))
)

// kevValMustParse takes a string and exits with a message if it cannot parse as KEY=VALUE
func keyValMustParse(str, message string) (string, string) {
	kv := strings.SplitN(str, "=", 2)
	if len(kv) != 2 {
		log.Fatalf("prometheus-to-cloudwatch: Error: %s", message)
	}
	return kv[0], kv[1]
}

// dimensionMatcherListMustParse takes a string and a flag name and exists with a message
// if it cannot parse as GLOB=dim1,dim2;GLOB2=dim3
func dimensionMatcherListMustParse(str, flag string) []MatcherWithStringSet {
	var matcherList []MatcherWithStringSet
	// split metric1=dim1,dim2;metric2=dim1
	//  into [
	//      metric1=dim1,dim2
	//      metric*=dim1
	// ]
	// then into [{ Matcher: "metric1": Set: [dim1, dim2] } , { Matcher: "metric_*": Set: [dim1] }]
	for _, sublist := range strings.Split(str, ";") {
		key, val := keyValMustParse(sublist, fmt.Sprintf("%s must be formatted as METRIC_NAME=DIM_LIST;...", flag))

		metricPattern, err := glob.Compile(key)
		if err != nil {
			log.Fatal(fmt.Errorf("prometheus-to-cloudwatch: Error: %s contains invalid glob pattern in '%s': %s", flag, key, err))
		}

		dims := strings.Split(val, ",")
		if len(dims) == 0 {
			log.Fatalf("prometheus-to-cloudwatch: Error: %s was not given dimensions to exclude for metric '%s'", flag, key)
		}
		g := MatcherWithStringSet{
			Matcher: metricPattern,
			Set:     stringSliceToSet(dims),
		}
		matcherList = append(matcherList, g)
	}
	return matcherList
}

// stringSliceToSet creates a "set" (a boolean map) from a slice of strings
func stringSliceToSet(slice []string) StringSet {
	boolMap := make(StringSet, len(slice))

	for i := range slice {
		boolMap[slice[i]] = true
	}

	return boolMap
}

func startHttpServer(ctx context.Context) {
	var metricsListenAddress = *listenAddress
	if metricsListenAddress == "" {
		metricsListenAddress = DEFAULT_LISTEN_ADDRESS
	}

	var metricsListenPath = *metricsPath
	if metricsListenPath == "" {
		metricsListenPath = DEFAULT_METRICS_PATH
	}

	httpServerExitDone := &sync.WaitGroup{}

	server := &http.Server{Addr: metricsListenAddress}
	http.Handle(metricsListenPath, promhttp.Handler())

	go func() {
		httpServerExitDone.Add(1)
		defer httpServerExitDone.Done()
		log.Println(fmt.Sprintf("prometheus-to-cloudwatch: Http server listening on %s", metricsListenAddress))
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln(fmt.Sprintf("prometheus-to-cloudwatch: Http server failed to listen on %s", metricsListenAddress), err)
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalln("prometheus-to-cloudwatch: Failed to gracefully stop Http server", err)
	}

	httpServerExitDone.Wait()
}

func main() {
	flag.Parse()

	if *cloudWatchNamespace == "" {
		flag.PrintDefaults()
		log.Fatal("prometheus-to-cloudwatch: Error: -cloudwatch_namespace or CLOUDWATCH_NAMESPACE required")
	}
	if *cloudWatchRegion == "" {
		flag.PrintDefaults()
		log.Fatal("prometheus-to-cloudwatch: Error: -cloudwatch_region or CLOUDWATCH_REGION required")
	}
	if *prometheusScrapeUrl == "" {
		flag.PrintDefaults()
		log.Fatal("prometheus-to-cloudwatch: Error: -prometheus_scrape_url or PROMETHEUS_SCRAPE_URL required")
	}
	if (*certPath != "" && *keyPath == "") || (*certPath == "" && *keyPath != "") {
		flag.PrintDefaults()
		log.Fatal("prometheus-to-cloudwatch: Error: when using SSL, both -prometheus_cert_path and -prometheus_key_path are required. If not using SSL, do not provide any of them")
	}

	var skipCertCheck = true
	var err error

	if *skipServerCertCheck != "" {
		if skipCertCheck, err = strconv.ParseBool(*skipServerCertCheck); err != nil {
			log.Fatal("prometheus-to-cloudwatch: Error: ", err)
		}
	}

	var additionalDimensions = map[string]string{}
	if *additionalDimension != "" {
		key, val := keyValMustParse(*additionalDimension, "-additionalDimension must be formatted as NAME=VALUE")
		additionalDimensions[key] = val
	}

	var replaceDims = map[string]string{}
	if *replaceDimensions != "" {
		kvs := strings.Split(*replaceDimensions, ",")
		if len(kvs) > 0 {
			for _, rd := range kvs {
				key, val := keyValMustParse(rd, "-replaceDimensions must be formatted as NAME=VALUE,...")
				replaceDims[key] = val
			}
		}
	}

	var includeMetricsList []glob.Glob
	if *includeMetrics != "" {
		for _, pattern := range strings.Split(*includeMetrics, ",") {
			g, err := glob.Compile(pattern)
			if err != nil {
				log.Fatal(fmt.Errorf("prometheus-to-cloudwatch: Error: -include_metrics contains invalid glob pattern in '%s': %s", pattern, err))
			}
			includeMetricsList = append(includeMetricsList, g)
		}
	}

	var excludeMetricsList []glob.Glob
	if *excludeMetrics != "" {
		for _, pattern := range strings.Split(*excludeMetrics, ",") {
			g, err := glob.Compile(pattern)
			if err != nil {
				log.Fatal(fmt.Errorf("prometheus-to-cloudwatch: Error: -exclude_metrics contains invalid glob pattern in '%s': %s", pattern, err))
			}
			excludeMetricsList = append(excludeMetricsList, g)
		}
	}

	var excludeDimensionsForMetricsList []MatcherWithStringSet
	if *excludeDimensionsForMetrics != "" {
		excludeDimensionsForMetricsList = dimensionMatcherListMustParse(*excludeDimensionsForMetrics, "-exclude_dimensions_for_metrics")
	}

	var includeDimensionsForMetricsList []MatcherWithStringSet
	if *includeDimensionsForMetrics != "" {
		includeDimensionsForMetricsList = dimensionMatcherListMustParse(*includeDimensionsForMetrics, "-include_dimensions_for_metrics")
	}

	config := &Config{
		CloudWatchNamespace:           *cloudWatchNamespace,
		CloudWatchRegion:              *cloudWatchRegion,
		PrometheusScrapeUrl:           *prometheusScrapeUrl,
		PrometheusCertPath:            *certPath,
		PrometheusKeyPath:             *keyPath,
		PrometheusSkipServerCertCheck: skipCertCheck,
		AwsAccessKeyId:                *awsAccessKeyId,
		AwsSecretAccessKey:            *awsSecretAccessKey,
		AwsSessionToken:               *awsSessionToken,
		AdditionalDimensions:          additionalDimensions,
		ReplaceDimensions:             replaceDims,
		IncludeMetrics:                includeMetricsList,
		ExcludeMetrics:                excludeMetricsList,
		ExcludeDimensionsForMetrics:   excludeDimensionsForMetricsList,
		IncludeDimensionsForMetrics:   includeDimensionsForMetricsList,
		ForceHighRes:                  *forceHighRes,
	}

	if *prometheusScrapeInterval != "" {
		interval, err := strconv.Atoi(*prometheusScrapeInterval)
		if err != nil {
			log.Fatal("prometheus-to-cloudwatch: error parsing 'prometheus_scrape_interval': ", err)
		}
		config.CloudWatchPublishInterval = time.Duration(interval) * time.Second
	}

	if *cloudWatchPublishTimeout != "" {
		timeout, err := strconv.Atoi(*cloudWatchPublishTimeout)
		if err != nil {
			log.Fatal("prometheus-to-cloudwatch: error parsing 'cloudwatch_publish_timeout': ", err)
		}
		config.CloudWatchPublishTimeout = time.Duration(timeout) * time.Second
	}

	bridge, err := NewBridge(config)

	if err != nil {
		log.Fatal("prometheus-to-cloudwatch: Error: ", err)
	}

	log.Println("prometheus-to-cloudwatch: Starting prometheus-to-cloudwatch bridge")

	ctx := context.Background()
	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(signals)
		cancel()
	}()
	go func() {
		select {
		case <-signals:
			cancel()
		case <-ctx.Done():
		}
	}()

	startHttpServer(ctx)
	bridge.Run(ctx)
}
