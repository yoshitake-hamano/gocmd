# File: Makefile - last edit
# yoshitake 16-Nov-2019

CXX        := g++
LD_LIBRARY :=
INCLUDE    :=
CPPFLAGS   := -Wall -g -O0 $(INCLUDE) $(LD_LIBRARY)
TARGETS    := $(basename $(wildcard *.c))
CW         := ../../bin/cw

all: $(TARGETS)


etags:
	find . -name "*.h" -o -name "*.c" -o -name "*.hpp" -o -name "*.cpp" | xargs etags

test: all
	$(MAKE) test-sucess
	$(MAKE) test-fails
	echo pass all tests

test-sucess: all
	$(CW) -i example -black blacklist.regexp -white whitelist.regexp
	$(CW) -i example -black blacklist.regexp -white whitelist.regexp -new_pass_list=-
	$(CW) -i example -black blacklist.regexp -white whitelist.regexp -result=-
	$(CW) -i example -black blacklist.regexp -white whitelist.regexp -new_pass_list=newPassList.txt
	$(CW) -i example -black blacklist.regexp -white whitelist.regexp -result=result.txt
	$(CW) -i example -black blacklist.regexp -white whitelist.regexp -pass=passList.txt

test-fails: all
	! $(CW) -i nofile -black blacklist.regexp -white whitelist.regexp
	! $(CW) -i example -black nofile -white whitelist.regexp
	! $(CW) -i example -black blacklist.regexp -white nofile
	! $(CW) -i example -black blacklist.regexp -white whitelist.regexp -ignore=nofile

clean:
	$(RM) $(TARGETS)
	$(RM) *~
	$(RM) newPassList.txt result.txt

# Log
# 16-Nov-2019 yoshitake Created.
