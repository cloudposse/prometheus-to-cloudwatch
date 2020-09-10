FROM golang:1.14.3 as builder
RUN mkdir -p /go/src/github.com/cloudposse/prometheus-to-cloudwatch
WORKDIR /go/src/github.com/cloudposse/prometheus-to-cloudwatch
COPY . .
RUN go get && CGO_ENABLED=0 go build -v -o "./dist/bin/prometheus-to-cloudwatch" *.go


FROM alpine:3.8
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/cloudposse/prometheus-to-cloudwatch/dist/bin/prometheus-to-cloudwatch /usr/bin/prometheus-to-cloudwatch
ENV PATH $PATH:/usr/bin
ENTRYPOINT ["prometheus-to-cloudwatch"]
