package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

type Finder struct {
	regexps []*regexp.Regexp
}

func NewFinder(searchWords []string) *Finder {
	regexps := make([]*regexp.Regexp, 0, len(searchWords))
	for _, word := range searchWords {
		r := regexp.MustCompile(word)
		regexps = append(regexps, r)
	}
	return &Finder{
		regexps: regexps,
	}
}

func (b *Finder) Find(src []byte, fn func([]byte)) []byte {
	for _, r := range b.regexps {
		for _, match := range r.FindAll(src, len(src)) {
			fn(match)
		}
	}
	return src
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func readRegexps(filename string) ([]string, error) {
	regexps := make([]string, 0)
	fp, err := os.Open(filename)
	defer fp.Close()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		regexps = append(regexps, scanner.Text())
	}
	return regexps, nil
}

func main() {
	var (
		inputfile  = flag.String("i", "", "input file")
		regexpfile = flag.String("b", "", "regexp file(blacklist)")
	)
	flag.Parse()

	filedata, err := ioutil.ReadFile(*inputfile)
	check(err)

	regexps, err := readRegexps(*regexpfile)
	check(err)
	b := NewFinder(regexps)
	b.Find(filedata, func(match []byte) {
		fmt.Printf("%s: %s\n", *inputfile, string(match))
	})
}
