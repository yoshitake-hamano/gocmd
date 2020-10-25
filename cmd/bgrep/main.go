package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type Finder struct {
	blacklist []*regexp.Regexp
	whitelist []*regexp.Regexp
}

func compileRegexps(regexps []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(regexps))
	for _, reg := range regexps {
		r := regexp.MustCompile(reg)
		compiled = append(compiled, r)
	}
	return compiled
}

func (b *Finder) matchWhitelist(src []byte) bool {
	for _, wr := range b.whitelist {
		if wr.Match(src) {
			return true
		}
	}
	return false
}

func NewFinder(blacklist, whitelist []string) *Finder {
	return &Finder{
		blacklist: compileRegexps(blacklist),
		whitelist: compileRegexps(whitelist),
	}
}

func (b *Finder) Find(src []byte, fn func([]byte)) []byte {
	for _, br := range b.blacklist {
		for _, match := range br.FindAll(src, len(src)) {
			if b.matchWhitelist(match) {
				continue
			}
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
		return nil, fmt.Errorf("read regexps: %w", err)
	}

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		regexps = append(regexps, scanner.Text())
	}
	return regexps, nil
}

func main() {
	var (
		inputpath  = flag.String("i", "", "input path")
		blacklistfile = flag.String("b", "", "regexp file(blacklist)")
		whitelistfile = flag.String("w", "", "regexp file(whitelist)")
	)
	flag.Parse()

	blacklist, err := readRegexps(*blacklistfile)
	check(err)
	whitelist, err := readRegexps(*whitelistfile)
	check(err)

	err = filepath.Walk(*inputpath, func(path string, info os.FileInfo, err error) error {
		if ! info.Mode().IsRegular() {
			return err
		}

		filedata, err := ioutil.ReadFile(path)
		check(err)

		b := NewFinder(blacklist, whitelist)
		b.Find(filedata, func(match []byte) {
			fmt.Printf("%s: %s\n", path, string(match))
		})
		return err
	})
	check(err)
}
