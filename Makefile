SHELL = /bin/bash

.PHONY : go-build
go-build: 
	CGO_ENABLED=0 go build -v -o "./dist/bin/prometheus-to-cloudwatch" *.go
