<!-- This file was automatically generated by the `build-harness`. Make all changes to `README.yaml` and run `make readme` to rebuild this file. -->
[![README Header][readme_header_img]][readme_header_link]

[![Cloud Posse][logo]](https://cpco.io/homepage)

# prometheus-to-cloudwatch [![Build Status](https://travis-ci.org/cloudposse/prometheus-to-cloudwatch.svg?branch=master)](https://travis-ci.org/cloudposse/prometheus-to-cloudwatch) [![Latest Release](https://img.shields.io/github/release/cloudposse/prometheus-to-cloudwatch.svg)](https://github.com/cloudposse/prometheus-to-cloudwatch/releases/latest) [![Slack Community](https://slack.cloudposse.com/badge.svg)](https://slack.cloudposse.com)



Utility for scraping Prometheus metrics from a Prometheus client endpoint and publishing them to CloudWatch


---

This project is part of our comprehensive ["SweetOps"](https://cpco.io/sweetops) approach towards DevOps. 
[<img align="right" title="Share via Email" src="https://docs.cloudposse.com/images/ionicons/ios-email-outline-2.0.1-16x16-999999.svg"/>][share_email]
[<img align="right" title="Share on Google+" src="https://docs.cloudposse.com/images/ionicons/social-googleplus-outline-2.0.1-16x16-999999.svg" />][share_googleplus]
[<img align="right" title="Share on Facebook" src="https://docs.cloudposse.com/images/ionicons/social-facebook-outline-2.0.1-16x16-999999.svg" />][share_facebook]
[<img align="right" title="Share on Reddit" src="https://docs.cloudposse.com/images/ionicons/social-reddit-outline-2.0.1-16x16-999999.svg" />][share_reddit]
[<img align="right" title="Share on LinkedIn" src="https://docs.cloudposse.com/images/ionicons/social-linkedin-outline-2.0.1-16x16-999999.svg" />][share_linkedin]
[<img align="right" title="Share on Twitter" src="https://docs.cloudposse.com/images/ionicons/social-twitter-outline-2.0.1-16x16-999999.svg" />][share_twitter]




It's 100% Open Source and licensed under the [APACHE2](LICENSE).











## Screenshots


![kube-state-metrics-to-cloudwatch](images/kube-state-metrics-to-cloudwatch.png)
*kube-state-metrics to CloudWatch*



## Usage




__NOTE__: The module accepts parameters as command-line arguments or as ENV variables (or any combination of command-line arguments and ENV vars).
Command-line arguments take precedence over ENV vars


| Command-line argument          | ENV var                        | Description                                                                                                                                                              |
|--------------------------------|--------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| aws_access_key_id              | AWS_ACCESS_KEY_ID              | AWS access key Id with permissions to publish CloudWatch metrics                                                                                                         |
| aws_secret_access_key          | AWS_SECRET_ACCESS_KEY          | AWS secret access key with permissions to publish CloudWatch metrics                                                                                                     |
| cloudwatch_namespace           | CLOUDWATCH_NAMESPACE           | CloudWatch Namespace                                                                                                                                                     |
| cloudwatch_region              | CLOUDWATCH_REGION              | CloudWatch AWS Region                                                                                                                                                    |
| cloudwatch_publish_timeout     | CLOUDWATCH_PUBLISH_TIMEOUT     | CloudWatch publish timeout in seconds                                                                                                                                    |
| prometheus_scrape_interval     | PROMETHEUS_SCRAPE_INTERVAL     | Prometheus scrape interval in seconds                                                                                                                                    |
| prometheus_scrape_url          | PROMETHEUS_SCRAPE_URL          | The URL to scrape Prometheus metrics from                                                                                                                                |
| cert_path                      | CERT_PATH                      | Path to SSL Certificate file (when using SSL for `prometheus_scrape_url`)                                                                                                |
| keyPath                        | KEY_PATH                       | Path to Key file (when using SSL for `prometheus_scrape_url`)                                                                                                            |
| accept_invalid_cert            | ACCEPT_INVALID_CERT            | Accept any certificate during TLS handshake. Insecure, use only for testing                                                                                              |
| additional_dimension           | ADDITIONAL_DIMENSION           | Additional dimension specified by NAME=VALUE                                                                                                                             |
| replace_dimensions             | REPLACE_DIMENSIONS             | Replace dimensions specified by NAME=VALUE,...                                                                                                                           |
| include_metrics                | INCLUDE_METRICS                | Only publish the specified metrics (comma-separated list of glob patterns)                                                                                               |
| exclude_metrics                | EXCLUDE_METRICS                | Never publish the specified metrics (comma-separated list of glob patterns)                                                                                              |
| exclude_dimensions_for_metrics | EXCLUDE_DIMENSIONS_FOR_METRICS | Dimensions to exclude for metrics (semi-colon-separated key values of comma-separated dimensions of METRIC=dim1,dim2;, e.g. 'flink_jobmanager=job,host;zk_up=host,pod;') |


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
export CERT_PATH=""
export KEY_PATH=""
export ACCEPT_INVALID_CERT=true
# Optionally, restrict the subset of metrics to be exported to CloudWatch
# export INCLUDE_METRICS='jvm_*'
# export EXCLUDE_METRICS='jvm_memory_*,jvm_buffer_*'
# export EXCLUDE_DIMENSIONS_FOR_METRICS='jvm_memory_*=pod;jvm_buffer=job,pod'

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
        -e CERT_PATH="" \
        -e KEY_PATH="" \
        -e ACCEPT_INVALID_CERT=true \
        -e INCLUDE_METRICS="" \
        -e EXCLUDE_METRICS="" \
        -e EXCLUDE_DIMENSIONS_FOR_METRICS="" \
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

Copy the name of the `kube-state-metrics` service (e.g. `gauche-turtle-kube-state-metrics`) into the ENV var `PROMETHEUS_SCRAPE_URL` in [values.yaml](chart/values.yaml).

![kube-state-metrics-service](images/kube-state-metrics-service.png)

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





## Share the Love 

Like this project? Please give it a ★ on [our GitHub](https://github.com/cloudposse/prometheus-to-cloudwatch)! (it helps us **a lot**) 

Are you using this project or any of our other projects? Consider [leaving a testimonial][testimonial]. =)


## Related Projects

Check out these related projects.

- [Prometheus Operator](https://github.com/cloudposse/prometheus-operator) - Prometheus Operator creates/configures/manages Prometheus clusters atop Kubernetes
- [terraform-aws-cloudwatch-logs](https://github.com/cloudposse/terraform-aws-cloudwatch-logs) - Terraform Module to Provide a CloudWatch Logs Endpoint
- [terraform-aws-ecs-web-app](https://github.com/cloudposse/terraform-aws-ecs-web-app) - Terraform module that implements a web app on ECS and supports autoscaling, CI/CD, monitoring, ALB integration, and much more.



## Help

**Got a question?**

File a GitHub [issue](https://github.com/cloudposse/prometheus-to-cloudwatch/issues), send us an [email][email] or join our [Slack Community][slack].

[![README Commercial Support][readme_commercial_support_img]][readme_commercial_support_link]

## Commercial Support

Work directly with our team of DevOps experts via email, slack, and video conferencing. 

We provide [*commercial support*][commercial_support] for all of our [Open Source][github] projects. As a *Dedicated Support* customer, you have access to our team of subject matter experts at a fraction of the cost of a full-time engineer. 

[![E-Mail](https://img.shields.io/badge/email-hello@cloudposse.com-blue.svg)][email]

- **Questions.** We'll use a Shared Slack channel between your team and ours.
- **Troubleshooting.** We'll help you triage why things aren't working.
- **Code Reviews.** We'll review your Pull Requests and provide constructive feedback.
- **Bug Fixes.** We'll rapidly work to fix any bugs in our projects.
- **Build New Terraform Modules.** We'll [develop original modules][module_development] to provision infrastructure.
- **Cloud Architecture.** We'll assist with your cloud strategy and design.
- **Implementation.** We'll provide hands-on support to implement our reference architectures. 




## Slack Community

Join our [Open Source Community][slack] on Slack. It's **FREE** for everyone! Our "SweetOps" community is where you get to talk with others who share a similar vision for how to rollout and manage infrastructure. This is the best place to talk shop, ask questions, solicit feedback, and work together as a community to build totally *sweet* infrastructure.

## Newsletter

Signup for [our newsletter][newsletter] that covers everything on our technology radar.  Receive updates on what we're up to on GitHub as well as awesome new projects we discover. 

## Contributing

### Bug Reports & Feature Requests

Please use the [issue tracker](https://github.com/cloudposse/prometheus-to-cloudwatch/issues) to report any bugs or file feature requests.

### Developing

If you are interested in being a contributor and want to get involved in developing this project or [help out](https://cpco.io/help-out) with our other projects, we would love to hear from you! Shoot us an [email][email].

In general, PRs are welcome. We follow the typical "fork-and-pull" Git workflow.

 1. **Fork** the repo on GitHub
 2. **Clone** the project to your own machine
 3. **Commit** changes to your own branch
 4. **Push** your work back up to your fork
 5. Submit a **Pull Request** so that we can review your changes

**NOTE:** Be sure to merge the latest changes from "upstream" before making a pull request!


## Copyright

Copyright © 2017-2019 [Cloud Posse, LLC](https://cpco.io/copyright)



## License 

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) 

See [LICENSE](LICENSE) for full details.

    Licensed to the Apache Software Foundation (ASF) under one
    or more contributor license agreements.  See the NOTICE file
    distributed with this work for additional information
    regarding copyright ownership.  The ASF licenses this file
    to you under the Apache License, Version 2.0 (the
    "License"); you may not use this file except in compliance
    with the License.  You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing,
    software distributed under the License is distributed on an
    "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
    KIND, either express or implied.  See the License for the
    specific language governing permissions and limitations
    under the License.









## Trademarks

All other trademarks referenced herein are the property of their respective owners.

## About

This project is maintained and funded by [Cloud Posse, LLC][website]. Like it? Please let us know by [leaving a testimonial][testimonial]!

[![Cloud Posse][logo]][website]

We're a [DevOps Professional Services][hire] company based in Los Angeles, CA. We ❤️  [Open Source Software][we_love_open_source].

We offer [paid support][commercial_support] on all of our projects.  

Check out [our other projects][github], [follow us on twitter][twitter], [apply for a job][jobs], or [hire us][hire] to help with your cloud strategy and implementation.



### Contributors

|  [![Erik Osterman][osterman_avatar]][osterman_homepage]<br/>[Erik Osterman][osterman_homepage] | [![Andriy Knysh][aknysh_avatar]][aknysh_homepage]<br/>[Andriy Knysh][aknysh_homepage] | [![Igor Rodionov][goruha_avatar]][goruha_homepage]<br/>[Igor Rodionov][goruha_homepage] | [![yufukui-m][yufukui-m_avatar]][yufukui-m_homepage]<br/>[yufukui-m][yufukui-m_homepage] | [![Satadru Biswas][sbiswas-suplari_avatar]][sbiswas-suplari_homepage]<br/>[Satadru Biswas][sbiswas-suplari_homepage] | [![Austin ce][austince_avatar]][austince_homepage]<br/>[Austin ce][austince_homepage] |
|---|---|---|---|---|---|

  [osterman_homepage]: https://github.com/osterman
  [osterman_avatar]: https://img.cloudposse.com/150x150/https://github.com/osterman.png
  [aknysh_homepage]: https://github.com/aknysh
  [aknysh_avatar]: https://img.cloudposse.com/150x150/https://github.com/aknysh.png
  [goruha_homepage]: https://github.com/goruha
  [goruha_avatar]: https://img.cloudposse.com/150x150/https://github.com/goruha.png
  [yufukui-m_homepage]: https://github.com/yufukui-m
  [yufukui-m_avatar]: https://img.cloudposse.com/150x150/https://github.com/yufukui-m.png
  [sbiswas-suplari_homepage]: https://github.com/sbiswas-suplari
  [sbiswas-suplari_avatar]: https://img.cloudposse.com/150x150/https://github.com/sbiswas-suplari.png
  [austince_homepage]: https://github.com/austince
  [austince_avatar]: https://img.cloudposse.com/150x150/https://github.com/austince.png



[![README Footer][readme_footer_img]][readme_footer_link]
[![Beacon][beacon]][website]

  [logo]: https://cloudposse.com/logo-300x69.svg
  [docs]: https://cpco.io/docs
  [website]: https://cpco.io/homepage
  [github]: https://cpco.io/github
  [jobs]: https://cpco.io/jobs
  [hire]: https://cpco.io/hire
  [slack]: https://cpco.io/slack
  [linkedin]: https://cpco.io/linkedin
  [twitter]: https://cpco.io/twitter
  [testimonial]: https://cpco.io/leave-testimonial
  [newsletter]: https://cpco.io/newsletter
  [email]: https://cpco.io/email
  [commercial_support]: https://cpco.io/commercial-support
  [we_love_open_source]: https://cpco.io/we-love-open-source
  [module_development]: https://cpco.io/module-development
  [terraform_modules]: https://cpco.io/terraform-modules
  [readme_header_img]: https://cloudposse.com/readme/header/img?repo=cloudposse/prometheus-to-cloudwatch
  [readme_header_link]: https://cloudposse.com/readme/header/link?repo=cloudposse/prometheus-to-cloudwatch
  [readme_footer_img]: https://cloudposse.com/readme/footer/img?repo=cloudposse/prometheus-to-cloudwatch
  [readme_footer_link]: https://cloudposse.com/readme/footer/link?repo=cloudposse/prometheus-to-cloudwatch
  [readme_commercial_support_img]: https://cloudposse.com/readme/commercial-support/img?repo=cloudposse/prometheus-to-cloudwatch
  [readme_commercial_support_link]: https://cloudposse.com/readme/commercial-support/link?repo=cloudposse/prometheus-to-cloudwatch
  [share_twitter]: https://twitter.com/intent/tweet/?text=prometheus-to-cloudwatch&url=https://github.com/cloudposse/prometheus-to-cloudwatch
  [share_linkedin]: https://www.linkedin.com/shareArticle?mini=true&title=prometheus-to-cloudwatch&url=https://github.com/cloudposse/prometheus-to-cloudwatch
  [share_reddit]: https://reddit.com/submit/?url=https://github.com/cloudposse/prometheus-to-cloudwatch
  [share_facebook]: https://facebook.com/sharer/sharer.php?u=https://github.com/cloudposse/prometheus-to-cloudwatch
  [share_googleplus]: https://plus.google.com/share?url=https://github.com/cloudposse/prometheus-to-cloudwatch
  [share_email]: mailto:?subject=prometheus-to-cloudwatch&body=https://github.com/cloudposse/prometheus-to-cloudwatch
  [beacon]: https://ga-beacon.cloudposse.com/UA-76589703-4/cloudposse/prometheus-to-cloudwatch?pixel&cs=github&cm=readme&an=prometheus-to-cloudwatch
