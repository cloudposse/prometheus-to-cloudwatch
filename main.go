package main

import (
	"flag"
	"fmt"
	"github.com/gobwas/glob"
	"golang.org/x/net/context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	awsAccessKeyId           = flag.String("aws_access_key_id", os.Getenv("AWS_ACCESS_KEY_ID"), "AWS access key Id with permissions to publish CloudWatch metrics")
	awsSecretAccessKey       = flag.String("aws_secret_access_key", os.Getenv("AWS_SECRET_ACCESS_KEY"), "AWS secret access key with permissions to publish CloudWatch metrics")
	cloudWatchNamespace      = flag.String("cloudwatch_namespace", os.Getenv("CLOUDWATCH_NAMESPACE"), "CloudWatch Namespace")
	cloudWatchRegion         = flag.String("cloudwatch_region", os.Getenv("CLOUDWATCH_REGION"), "CloudWatch Region")
	cloudWatchPublishTimeout = flag.String("cloudwatch_publish_timeout", os.Getenv("CLOUDWATCH_PUBLISH_TIMEOUT"), "CloudWatch publish timeout in seconds")
	prometheusScrapeInterval = flag.String("prometheus_scrape_interval", os.Getenv("PROMETHEUS_SCRAPE_INTERVAL"), "Prometheus scrape interval in seconds")
	prometheusScrapeUrl      = flag.String("prometheus_scrape_url", os.Getenv("PROMETHEUS_SCRAPE_URL"), "Prometheus scrape URL")
	certPath                 = flag.String("cert_path", os.Getenv("CERT_PATH"), "Path to SSL Certificate file (when using SSL for `prometheus_scrape_url`)")
	keyPath                  = flag.String("key_path", os.Getenv("KEY_PATH"), "Path to Key file (when using SSL for `prometheus_scrape_url`)")
	skipServerCertCheck      = flag.String("accept_invalid_cert", os.Getenv("ACCEPT_INVALID_CERT"), "Accept any certificate during TLS handshake. Insecure, use only for testing")
	additionalDimension      = flag.String("additional_dimension", os.Getenv("ADDITIONAL_DIMENSION"), "Additional dimension specified by NAME=VALUE")
	replaceDimensions        = flag.String("replace_dimensions", os.Getenv("REPLACE_DIMENSIONS"), "replace dimensions specified by NAME=VALUE,...")
	includeMetrics           = flag.String("include_metrics", os.Getenv("INCLUDE_METRICS"), "Only publish the specified metrics (comma-separated list of glob patterns, e.g. 'up,http_*')")
	excludeMetrics           = flag.String("exclude_metrics", os.Getenv("EXCLUDE_METRICS"), "Never publish the specified metrics (comma-separated list of glob patterns, e.g. 'tomcat_*')")
)

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
		kv := strings.SplitN(*additionalDimension, "=", 2)
		if len(kv) != 2 {
			log.Fatal("prometheus-to-cloudwatch: Error: -additionalDimension must be formatted as NAME=VALUE")
		}
		additionalDimensions[kv[0]] = kv[1]
	}

	var replaceDims = map[string]string{}
	if *replaceDimensions != "" {
		kvs := strings.Split(*replaceDimensions, ",")
		if len(kvs) > 0 {
			for _, rd := range kvs {
				kv := strings.SplitN(rd, "=", 2)
				if len(kv) != 2 {
					log.Fatal("prometheus-to-cloudwatch: Error: -replaceDimensions must be formatted as NAME=VALUE,...")
				}
				replaceDims[kv[0]] = kv[1]
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

	config := &Config{
		CloudWatchNamespace:           *cloudWatchNamespace,
		CloudWatchRegion:              *cloudWatchRegion,
		PrometheusScrapeUrl:           *prometheusScrapeUrl,
		PrometheusCertPath:            *certPath,
		PrometheusKeyPath:             *keyPath,
		PrometheusSkipServerCertCheck: skipCertCheck,
		AwsAccessKeyId:                *awsAccessKeyId,
		AwsSecretAccessKey:            *awsSecretAccessKey,
		AdditionalDimensions:          additionalDimensions,
		ReplaceDimensions:             replaceDims,
		IncludeMetrics:                includeMetricsList,
		ExcludeMetrics:                excludeMetricsList,
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

	fmt.Println("prometheus-to-cloudwatch: Starting prometheus-to-cloudwatch bridge")
	bridge.Run(context.Background())
}
