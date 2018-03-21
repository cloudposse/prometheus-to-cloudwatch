SHELL = /bin/bash

PATH:=$(PATH):$(GOPATH)/bin

include $(shell curl --silent -o .build-harness "https://raw.githubusercontent.com/cloudposse/build-harness/master/templates/Makefile.build-harness"; echo .build-harness)


.PHONY : go-get
go-get:
	go get


.PHONY : go-build
go-build: go-get
	CGO_ENABLED=0 go build -v -o "./dist/bin/prometheus-to-cloudwatch" *.go
