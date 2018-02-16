# prometheus-to-cloudwatch Helm Chart

* Installs the [prometheus-to-cloudwatch agent](https://github.com/cloudposse/prometheus-to-cloudwatch).

## Installing the Chart

```bash
$ helm install cloudposse/prometheus-to-cloudwatch
```

## Configuration

| Parameter                           | Description                                             | Default                                     |
|-------------------------------------|---------------------------------------------------------|---------------------------------------------|
| `image.repository`                  | The image repository to pull from                       | cloudposse/prometheus-to-cloudwatch         |
