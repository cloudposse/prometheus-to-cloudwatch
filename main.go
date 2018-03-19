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
	cloudWatchNamespace           = flag.String("cloudwatch_namespace", os.Getenv("CLOUDWATCH_NAMESPACE"), "CloudWatch Namespace")
	cloudWatchRegion              = flag.String("cloudwatch_region", os.Getenv("CLOUDWATCH_REGION"), "CloudWatch Region")
	cloudWatchPublishTimeout      = flag.String("cloudwatch_publish_timeout", os.Getenv("CLOUDWATCH_PUBLISH_TIMEOUT"), "CloudWatch publish timeout in seconds")
	prometheusScrapeInterval      = flag.String("prometheus_scrape_interval", os.Getenv("PROMETHEUS_SCRAPE_INTERVAL"), "Prometheus scrape interval in seconds")
	prometheusScrapeUrl           = flag.String("prometheus_scrape_url", os.Getenv("PROMETHEUS_SCRAPE_URL"), "Prometheus scrape URL")
	prometheusCertPath            = flag.String("prometheus_cert_path", os.Getenv("PROMETHEUS_CERT_PATH"), "Path to Certificate file")
	prometheusKeyPath             = flag.String("prometheus_key_path", os.Getenv("PROMETHEUS_KEY_PATH"), "Path to Key file")
	prometheusSkipServerCertCheck = flag.String("prometheus_accept_invalid_cert", os.Getenv("PROMETHEUS_ACCEPT_INVALID_CERT"), "Accept any certificate during TLS handshake. Insecure, use only for testing")
)

func main() {
	flag.Parse()

	if *cloudWatchNamespace == "" {
		flag.PrintDefaults()
		log.Fatal("-cloudwatch_namespace or CLOUDWATCH_NAMESPACE required")
	}
	if *cloudWatchRegion == "" {
		flag.PrintDefaults()
		log.Fatal("-cloudwatch_region or CLOUDWATCH_REGION required")
	}
	if *prometheusScrapeUrl == "" {
		flag.PrintDefaults()
		log.Fatal("-prometheus_scrape_url or PROMETHEUS_SCRAPE_URL required")
	}
	if (*prometheusCertPath != "" && *prometheusKeyPath == "") || (*prometheusCertPath == "" && *prometheusKeyPath != "") {
		flag.PrintDefaults()
		log.Fatal("When using SSL, both -prometheus_cert_path and -prometheus_key_path are required. If not using SSL, do not provide any of them")
	}

	var skipServerCertCheck = true
	var err error

	if *prometheusSkipServerCertCheck != "" {
		if skipServerCertCheck, err = strconv.ParseBool(*prometheusSkipServerCertCheck); err != nil {
			log.Fatal("prometheus-to-cloudwatch: Error: ", err)
		}
	}

	config := &Config{
		CloudWatchNamespace:           *cloudWatchNamespace,
		CloudWatchRegion:              *cloudWatchRegion,
		PrometheusScrapeUrl:           *prometheusScrapeUrl,
		PrometheusCertPath:            *prometheusCertPath,
		PrometheusKeyPath:             *prometheusKeyPath,
		PrometheusSkipServerCertCheck: skipServerCertCheck,
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
