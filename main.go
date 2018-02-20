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
	cloudWatchNamespace       = flag.String("cloudwatch_namespace", os.Getenv("CLOUDWATCH_NAMESPACE"), "CloudWatch Namespace")
	cloudWatchRegion          = flag.String("cloudwatch_region", os.Getenv("CLOUDWATCH_REGION"), "CloudWatch Region")
	cloudWatchPublishInterval = flag.String("cloudwatch_publish_interval", os.Getenv("CLOUDWATCH_PUBLISH_INTERVAL"), "CloudWatch publish interval in seconds")
	cloudWatchPublishTimeout  = flag.String("cloudwatch_publish_timeout", os.Getenv("CLOUDWATCH_PUBLISH_TIMEOUT"), "CloudWatch publish timeout in seconds")
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

	config := &Config{
		CloudWatchNamespace: *cloudWatchNamespace,
		CloudWatchRegion:    *cloudWatchRegion,
	}

	if *cloudWatchPublishInterval != "" {
		interval, err := strconv.Atoi(*cloudWatchPublishInterval)
		if err != nil {
			log.Fatal("prometheus-to-cloudwatch: error parsing 'cloudwatch_publish_interval': ", err)
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
		log.Fatal("prometheus-to-cloudwatch: error: ", err)
	}

	bridge.Run(context.Background())
	fmt.Println("prometheus-to-cloudwatch: Started prometheus-to-cloudwatch bridge")
}
