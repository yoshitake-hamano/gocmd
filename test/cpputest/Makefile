
SRCS     := $(wildcard test_*.cpp)
TARGETS  := $(SRCS:%.cpp=%)
CPPFLAGS := -Wall -lcpputest
GPP      := g++

.PHONY: test
test: $(TARGETS)
	for cmd in $^; do \
		./$$cmd; \
	done

test_%: test_%.cpp
	$(GPP) $(CPPFLAGS) -o $@ $^
