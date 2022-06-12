# This is a regular comment, that will not be displayed

## Targets
## 

.PHONY: all
all: build-lambda build-normal ## Build all

.PHONY: build-lambda
build-lambda: ## Build lambda binary
	cd cmd/mindra; GOOS=linux go build -tags lambda -o ../../binlambda
	cd binlambda; zip mindra.zip mindra

.PHONY: build-normal
build-normal: ## Build normal binary
	go generate ./...
	go build -o bin ./...

.PHONY: deploy-lambda
deploy-lambda: build-lambda ## Deploy to AWS Lambda
	cd binlambda; aws lambda update-function-code --no-cli-pager --function-name mindra --zip-file fileb://mindra.zip

.PHONY: install
install: ## Install bin files
	cp bin/* $(HOME)/bin/

.PHONY: clean
clean: ## Clean
	$(RM) binlambda/*
	$(RM) bin/*
	$(MAKE) clean -C test/cpputest
	$(MAKE) clean -C test/cw
	$(MAKE) clean -C test/blackout
	$(RM) cmd/createmock/parser.go
	$(RM) cmd/createmock/y.output

.PHONY: test
test: all ## Test
	go test ./...
	cd test/scripts; ./testMain.sh
	$(MAKE) test -C test/cpputest
	$(MAKE) test -C test/cw
	$(MAKE) test -C test/blackout

.PHONY: benchmark
benchmark: all ## Benchmark
	cd cmd/cw; go test -bench=. -trace a.trace

help: ## Show this help.
	@sed -ne "/@sed/!s/## //p" $(MAKEFILE_LIST)
