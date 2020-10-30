package main

import (
	"bufio"
	"bytes"
	"debug/elf"
	"flag"
	"fmt"
	"io"
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

func isTokenable(b byte) bool {
	// see ascii table
	if 0x20 <= b && b <= 0x7e {
		return true
	}
	return b == '\t'
}

func split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		return 0, nil, nil
	}
	startOfToken := len(data)
	for i:=0; i<len(data); i++ {
		if isTokenable(data[i]) {
			startOfToken = i
			break
		}
	}
	for i:=startOfToken; i<len(data); i++ {
		if ! isTokenable(data[i]) {
			return i+1, data[startOfToken:i], nil
		}
	}
	return len(data), data[startOfToken:], nil
}

func (b *Finder) findBinary(path string, r io.Reader, fn func(path, keyword, text string)) error {
	scanner := bufio.NewScanner(r)
	scanner.Split(split)
	for scanner.Scan() {
		t := scanner.Text()
		if t == "" {
			continue
		}
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
	return nil
}

func (b *Finder) findElf(path string, r io.Reader, fn func(path, keyword, text string)) error {
	f, err := elf.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, section := range f.Sections {
		src, err := section.Data()
		if err != nil {
			return err
		}
		err = b.findBinary(path, bytes.NewReader(src), fn)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Finder) Find(path string, r io.Reader, fn func(path, keyword, text string)) error {
	r1 := bytes.NewBuffer(nil)
	r2 := io.TeeReader(r, r1)
	if b.findElf(path, r1, fn) == nil {
		return nil
	}
	return b.findBinary(path, r2, fn)
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
	t := strings.Trim(text, "\t,")
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
			r, err := os.Open(p)

			if err != nil {
				fmt.Print(err)
				r.Close()
				continue
			}
			b.Find(p, r, printer)
			r.Close()
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

		r, err := os.Open(path)
		defer r.Close()

		if err != nil {
			fmt.Print(err)
			return nil
		}
		b.Find(path, r, printer)
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
