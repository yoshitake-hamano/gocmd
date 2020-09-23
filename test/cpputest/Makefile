
HEADERS      := $(wildcard *.h)
MOCK_HEADERS := $(HEADERS:%.h=mock_%.cpp)
SRCS         := $(wildcard test_*.cpp)
TARGETS      := $(SRCS:%.cpp=%)
CPPFLAGS     := -g -Wall -lcpputest  -lcpputestext
GPP          := g++

.PHONY: test
test: $(TARGETS)
	for cmd in $^; do \
		./$$cmd; \
	done

.PHONY: mock
mock: $(MOCK_HEADERS)

.PHONY: clean
clean:
	$(RM) $(MOCK_HEADERS) $(TARGETS)

mock_%.cpp: %.h
	../../bin/createmock -file $^ > $@

test_%: test_%.cpp $(MOCK_HEADERS)
	$(GPP) $(CPPFLAGS) -o $@ $<
