# File: Makefile - last edit
# yoshitake 16-Nov-2019

CXX          := g++
LD_LIBRARY   :=
INCLUDE      :=
CPPFLAGS     := -Wall -g -O0 $(INCLUDE) $(LD_LIBRARY)
TARGETS      := $(basename $(wildcard *.c))
TARGETS_DSYM := $(TARGETS:%=%.dSYM)


all: $(TARGETS)


etags:
	find . -name "*.h" -o -name "*.c" -o -name "*.hpp" -o -name "*.cpp" | xargs etags

test: all
	# ../../bin/blackout -i example -o example.out -r blackout.regexp

clean:
	$(RM)  $(TARGETS)
	$(RM) -r $(TARGETS_DSYM)

# Log
# 16-Nov-2019 yoshitake Created.
