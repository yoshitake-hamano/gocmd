# File: Makefile - last edit
# yoshitake 16-Nov-2019

CXX        := g++
LD_LIBRARY :=
INCLUDE    :=
CPPFLAGS   := -Wall -g -O0 $(INCLUDE) $(LD_LIBRARY)
TARGETS    := $(basename $(wildcard *.c))
BGREP      := ../../bin/bgrep

all: $(TARGETS)


etags:
	find . -name "*.h" -o -name "*.c" -o -name "*.hpp" -o -name "*.cpp" | xargs etags

test: all
	$(BGREP) -i example -b blacklist.regexp -w whitelist.regexp

clean:
	rm -f  $(TARGETS)
	rm -f *~

# Log
# 16-Nov-2019 yoshitake Created.
