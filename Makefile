# This is a regular comment, that will not be displayed

## Targets
## 


.PHONY: all
all: ## Build all
	go generate ./...
	go build -o bin ./...

.PHONY: clean
clean: ## Clean
	$(RM) bin/*
	cd test/cpputest; $(MAKE) clean

.PHONY: test
test: all ## Test
	go test ./...
	cd test/scripts; ./testMain.sh
	$(MAKE) test -C test/cpputest
	$(MAKE) test -C test/cw

.PHONY: benchmark
benchmark: all ## Benchmark
	cd cmd/cw; go test ./... -bench=. -trace a.trace

help: ## Show this help.
	@sed -ne "/@sed/!s/## //p" $(MAKEFILE_LIST)
