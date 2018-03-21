# prometheus-to-cloudwatch [![Build Status](https://travis-ci.org/cloudposse/prometheus-to-cloudwatch.svg?branch=master)](https://travis-ci.org/cloudposse/prometheus-to-cloudwatch)

Utility for scraping Prometheus metrics from a Prometheus client endpoint and publishing them to CloudWatch


## Usage

__NOTE__: The module accepts parameters as command-line arguments or as ENV variables (or any combination of command-line arguments and ENV vars).
Command-line arguments take precedence over ENV vars


| Command-line argument        |  ENV var                     |  Description                                                                  |
|:-----------------------------|:-----------------------------|:------------------------------------------------------------------------------|
| aws_access_key_id            | AWS_ACCESS_KEY_ID            | AWS access key Id with permissions to publish CloudWatch metrics              |
| aws_secret_access_key        | AWS_SECRET_ACCESS_KEY        | AWS secret access key with permissions to publish CloudWatch metrics          |
| cloudwatch_namespace         | CLOUDWATCH_NAMESPACE         | CloudWatch Namespace                                                          |
| cloudwatch_region            | CLOUDWATCH_REGION            | CloudWatch AWS Region                                                         |
| cloudwatch_publish_timeout   | CLOUDWATCH_PUBLISH_TIMEOUT   | CloudWatch publish timeout in seconds                                         |
| prometheus_scrape_interval   | PROMETHEUS_SCRAPE_INTERVAL   | Prometheus scrape interval in seconds                                         |
| prometheus_scrape_url        | PROMETHEUS_SCRAPE_URL        | The URL to scrape Prometheus metrics from                                     |
| cert_path                    | CERT_PATH                    | Path to SSL Certificate file (when using SSL for `prometheus_scrape_url`)     |
| keyPath                      | KEY_PATH                     | Path to Key file (when using SSL for `prometheus_scrape_url`)                 |
| accept_invalid_cert          | ACCEPT_INVALID_CERT          | Accept any certificate during TLS handshake. Insecure, use only for testing   |


__NOTE__: If AWS credentials are not provided in the command-line arguments (`aws_access_key_id` and `aws_secret_access_key`)
or ENV variables (`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`),
the chain of credential providers will search for credentials in the shared credential file and EC2 Instance Roles.
This is useful when deploying the module in AWS on Kubernetes with [`kube2iam`](https://github.com/jtblin/kube2iam),
which will provide IAM credentials to containers running inside a Kubernetes cluster, allowing the module to assume an IAM Role with permissions
to publish metrics to CloudWatch.


## Examples

### Build Go program

```sh
go get

CGO_ENABLED=0 go build -v -o "./dist/bin/prometheus-to-cloudwatch" *.go
```


### Run locally

```sh
export AWS_ACCESS_KEY_ID=XXXXXXXXXXXXXXXXXXXXXXX
export AWS_SECRET_ACCESS_KEY=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
export CLOUDWATCH_NAMESPACE=kube-state-metrics
export CLOUDWATCH_REGION=us-east-1
export CLOUDWATCH_PUBLISH_TIMEOUT=5
export PROMETHEUS_SCRAPE_INTERVAL=30
export PROMETHEUS_SCRAPE_URL=http://xxxxxxxxxxxx:8080/metrics
export PROMETHEUS_CERT_PATH=""
export PROMETHEUS_KEY_PATH=""
export PROMETHEUS_ACCEPT_INVALID_CERT=true

./dist/bin/prometheus-to-cloudwatch
```


### Build Docker image
__NOTE__: it will download all `Go` dependencies and then build the program inside the container (see [`Dockerfile`](Dockerfile))


```sh
docker build --tag prometheus-to-cloudwatch  --no-cache=true .
```


### Run in a Docker container

```sh
docker run -i --rm \
        -e AWS_ACCESS_KEY_ID=XXXXXXXXXXXXXXXXXXXXXXX \
        -e AWS_SECRET_ACCESS_KEY=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX \
        -e CLOUDWATCH_NAMESPACE=kube-state-metrics \
        -e CLOUDWATCH_REGION=us-east-1 \
        -e CLOUDWATCH_PUBLISH_TIMEOUT=5 \
        -e PROMETHEUS_SCRAPE_INTERVAL=30 \
        -e PROMETHEUS_SCRAPE_URL=http://xxxxxxxxxxxx:8080/metrics \
        -e PROMETHEUS_CERT_PATH="" \
        -e PROMETHEUS_KEY_PATH="" \
        -e PROMETHEUS_ACCEPT_INVALID_CERT=true \
        prometheus-to-cloudwatch
```


### Run on Kubernetes

To run on `Kubernetes`, we will deploy two [`Helm`](https://helm.sh/) [charts](https://docs.helm.sh/developing_charts/)

1. [kube-state-metrics](https://github.com/kubernetes/charts/tree/master/stable/kube-state-metrics) - to generates metrics about the state of various objects inside the cluster, such as deployments, nodes and pods

2. [prometheus-to-cloudwatch](chart) - to scrape metrics from `kube-state-metrics` and publish them to CloudWatch

Install `kube-state-metrics` chart

```sh
helm install stable/kube-state-metrics
```

Find the running services

```sh
kubectl get services
```

![kube-state-metrics-service](images/kube-state-metrics-service.png)


Copy the name of the `kube-state-metrics` service (`gauche-turtle-kube-state-metrics`) into the ENV var `PROMETHEUS_SCRAPE_URL` in [values.yaml](chart/values.yaml).
It should look like this:

```sh
PROMETHEUS_SCRAPE_URL: "http://gauche-turtle-kube-state-metrics:8080/metrics"
```

Deploy `prometheus-to-cloudwatch` chart

```sh
cd chart
helm install .
```

`prometheus-to-cloudwatch` will start scraping the `/metrics` endpoint of the `kube-state-metrics` service and send the Prometheus metrics to CloudWatch


![kube-state-metrics-to-cloudwatch](images/kube-state-metrics-to-cloudwatch.png)


## Help

**Got a question?**

File a GitHub [issue](https://github.com/cloudposse/prometheus-to-cloudwatch/issues), send us an [email](mailto:hello@cloudposse.com) or reach out to us on [Gitter](https://gitter.im/cloudposse/).


## Contributing

### Bug Reports & Feature Requests

Please use the [issue tracker](https://github.com/cloudposse/prometheus-to-cloudwatch/issues) to report any bugs or file feature requests.

### Developing

If you are interested in being a contributor and want to get involved in developing `prometheus-to-cloudwatch`, we would love to hear from you! Shoot us an [email](mailto:hello@cloudposse.com).

In general, PRs are welcome. We follow the typical "fork-and-pull" Git workflow.

 1. **Fork** the repo on GitHub
 2. **Clone** the project to your own machine
 3. **Commit** changes to your own branch
 4. **Push** your work back up to your fork
 5. Submit a **Pull request** so that we can review your changes

**NOTE:** Be sure to merge the latest from "upstream" before making a pull request!


## License

[APACHE 2.0](LICENSE) © 2018 [Cloud Posse, LLC](https://cloudposse.com)

See [LICENSE](LICENSE) for full details.

    Licensed to the Apache Software Foundation (ASF) under one
    or more contributor license agreements.  See the NOTICE file
    distributed with this work for additional information
    regarding copyright ownership.  The ASF licenses this file
    to you under the Apache License, Version 2.0 (the
    "License"); you may not use this file except in compliance
    with the License.  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing,
    software distributed under the License is distributed on an
    "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
    KIND, either express or implied.  See the License for the
    specific language governing permissions and limitations
    under the License.


## About

`prometheus-to-cloudwatch` is maintained and funded by [Cloud Posse, LLC][website].

![Cloud Posse](https://cloudposse.com/logo-300x69.png)


Like it? Please let us know at <hello@cloudposse.com>

We love [Open Source Software](https://github.com/cloudposse/)!

See [our other projects][community]
or [hire us][hire] to help build your next cloud platform.

  [website]: https://cloudposse.com/
  [community]: https://github.com/cloudposse/
  [hire]: https://cloudposse.com/contact/


### Contributors

| [![Erik Osterman][erik_img]][erik_web]<br/>[Erik Osterman][erik_web] | [![Andriy Knysh][andriy_img]][andriy_web]<br/>[Andriy Knysh][andriy_web] |
|-------------------------------------------------------|------------------------------------------------------------------|

  [erik_img]: http://s.gravatar.com/avatar/88c480d4f73b813904e00a5695a454cb?s=144
  [erik_web]: https://github.com/osterman/
  [andriy_img]: https://avatars0.githubusercontent.com/u/7356997?v=4&u=ed9ce1c9151d552d985bdf5546772e14ef7ab617&s=144
  [andriy_web]: https://github.com/aknysh/
