package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"log"
	"os"
	"strconv"
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

	config := &Config{
		CloudWatchNamespace:           *cloudWatchNamespace,
		CloudWatchRegion:              *cloudWatchRegion,
		PrometheusScrapeUrl:           *prometheusScrapeUrl,
		PrometheusCertPath:            *certPath,
		PrometheusKeyPath:             *keyPath,
		PrometheusSkipServerCertCheck: skipCertCheck,
		AwsAccessKeyId:                *awsAccessKeyId,
		AwsSecretAccessKey:            *awsSecretAccessKey,
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
