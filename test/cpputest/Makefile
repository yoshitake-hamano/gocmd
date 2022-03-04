# This is a regular comment, that will not be displayed

## Targets
## 


HEADERS      := $(wildcard *.h)
MOCK_HEADERS := $(HEADERS:%.h=mock_%.cpp)
SRCS         := $(wildcard test_*.cpp)
TARGETS      := $(SRCS:%.cpp=%)
TARGETS_DSYM := $(TARGETS:%=%.dSYM)
CPPFLAGS     := -g -Wall -lcpputest  -lcpputestext
GPP          := g++

.PHONY: test
test: $(TARGETS) ## Test
	for cmd in $^; do \
		./$$cmd -v; \
	done

.PHONY: mock
mock: $(MOCK_HEADERS) ## Build mock

.PHONY: clean
clean: ## Clean
	$(RM) $(MOCK_HEADERS) $(TARGETS)
	$(RM) -r $(TARGETS_DSYM)

mock_%.cpp: %.h
	../../bin/createmock -vervose -file $^ > $@

test_%: test_%.cpp $(MOCK_HEADERS)
	$(GPP) $(CPPFLAGS) -o $@ $<

help: ## Show this help.
	@sed -ne "/@sed/!s/## //p" $(MAKEFILE_LIST)
