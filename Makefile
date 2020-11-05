SHELL = /bin/bash

include $(shell curl --silent -o .build-harness "https://raw.githubusercontent.com/cloudposse/build-harness/master/templates/Makefile.build-harness"; echo .build-harness)

.PHONY : go-build
go-build: 
	CGO_ENABLED=0 go build -v -o "./dist/bin/prometheus-to-cloudwatch" *.go

## Lint code
lint: $(GO) vet
	$(call assert-set,GO)
	find . ! -path "*/vendor/*" ! -path "*/.glide/*" -type f -name '*.go' | xargs -n 1 golint

## Vet code
vet: $(GO)
	$(call assert-set,GO)
	find . ! -path "*/vendor/*" ! -path "*/.glide/*" -type f -name '*.go' | xargs $(GO) vet -v


