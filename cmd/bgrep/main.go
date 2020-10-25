package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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

func (b *Finder) matchWhitelist(src string) (bool, *regexp.Regexp) {
	for _, wr := range b.whitelist {
		if wr.MatchString(src) {
			return true, wr
		}
	}
	return false, nil
}

func (b *Finder) matchBlacklist(src string) (bool, *regexp.Regexp) {
	for _, wr := range b.blacklist {
		if wr.MatchString(src) {
			return true, wr
		}
	}
	return false, nil
}

func NewFinder(blacklist, whitelist []string) *Finder {
	return &Finder{
		blacklist: compileRegexps(blacklist),
		whitelist: compileRegexps(whitelist),
	}
}

func (b *Finder) Find(path string, fn func(path, keyword, text string)) {
	r, err := os.Open(path)
	defer r.Close()

	if err != nil {
		fmt.Print(err)
		return
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		t := scanner.Text()
		match := true
		var keyword *regexp.Regexp
		if match, keyword = b.matchBlacklist(t); !match {
			continue
		}
		if match, _ = b.matchWhitelist(t); match {
			continue
		}
		fn(path, keyword.String(), t)
	}
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

func printMatchString(path, keyword, text string) {
	t := strings.TrimFunc(text, func(r rune) bool {
 		switch r {
 		case rune('\r'):
 		case rune('\n'):
 		case rune('\t'):
 			return true
 		}
		// unicode.In(r, unicode.N, unicode.L, unicode.M)
		return false
	})
	fmt.Printf("%s,%s,%s\n", path, keyword, t)
}

func printDummyMatchString(path, keyword, text string) {
}

func mainImplUsingGoroutine(blacklist, whitelist []string, inputpath string) error {
	b := NewFinder(blacklist, whitelist)
	printer := printMatchString
	if *silent {
		printer = printDummyMatchString
	}

	ch := make(chan string)
	wg := sync.WaitGroup{}
	worker := func() {
		defer wg.Done()
		for p := range ch {
			b.Find(p, printer)
		}
	}
	const sizeOfGorotine = 10
	wg.Add(sizeOfGorotine)
	for i:=0; i<sizeOfGorotine; i++ {
		go worker()
	}
	err := filepath.Walk(inputpath, func(path string, info os.FileInfo, err error) error {
		if ! info.Mode().IsRegular() {
			return err
		}
		ch <- path
		return nil
	})
	close(ch)
	wg.Wait()
	return err
}

func mainImplStanderd(blacklist, whitelist []string, inputpath string) error {
	b := NewFinder(blacklist, whitelist)
	printer := printMatchString
	if *silent {
		printer = printDummyMatchString
	}
	err := filepath.Walk(inputpath, func(path string, info os.FileInfo, err error) error {
		if ! info.Mode().IsRegular() {
			return err
		}

		b.Find(path, printer)
		return nil
	})
	return err
}

var silent = flag.Bool("s", false, "silent(for benchmark)")

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

	err = mainImplStanderd(blacklist, whitelist, *inputpath)
	check(err)
}
