package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
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

func mainImplUsingGoroutine(blacklist, whitelist []string, inputpath string) error {
	b := NewFinder(blacklist, whitelist)
	ch := make(chan string)
	wg := sync.WaitGroup{}
	fn := func() {
		defer wg.Done()
		for p := range ch {
			filedata, err := ioutil.ReadFile(p)
			if err != nil {
				fmt.Print(err)
				return
			}

			b.Find(filedata, func(match []byte) {
				if ! *silent {
					fmt.Printf("%s: %s\n", p, string(match))
				}
			})
		}
	}
	const sizeOfGorotine = 10
	wg.Add(sizeOfGorotine)
	for i:=0; i<sizeOfGorotine; i++ {
		go fn()
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
	err := filepath.Walk(inputpath, func(path string, info os.FileInfo, err error) error {
		if ! info.Mode().IsRegular() {
			return err
		}

		filedata, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		b.Find(filedata, func(match []byte) {
			if ! *silent {
				fmt.Printf("%s: %s\n", path, string(match))
			}
		})
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
