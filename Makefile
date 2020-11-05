SHELL = /bin/bash

include $(shell curl --silent -o .build-harness "https://raw.githubusercontent.com/cloudposse/build-harness/master/templates/Makefile.build-harness"; echo .build-harness)

.PHONY : go-build
go-build: 
	CGO_ENABLED=0 go build -v -o "./dist/bin/prometheus-to-cloudwatch" *.go
