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
	"sync"
)

type Finder struct {
	blacklist []*regexp.Regexp
	whitelist []*regexp.Regexp
}

type ResultWriter interface {
	Write(path, filetype, section, keyword, text string)
}

type ResultWriterImpl struct {
	w io.Writer
}

type DummyResultWriter struct {
}

func compileRegexps(regexps []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(regexps))
	for _, reg := range regexps {
		r := regexp.MustCompile(reg)
		compiled = append(compiled, r)
	}
	return compiled
}

func matchRegexps(str string, regexps []*regexp.Regexp) (bool, *regexp.Regexp) {
	for _, reg := range regexps {
		if reg.MatchString(str) {
			return true, reg
		}
	}
	return false, nil
}

func matchAllRegexps(str string, regexps []*regexp.Regexp) []*regexp.Regexp {
	r := make([]*regexp.Regexp, 0)
	for _, reg := range regexps {
		if reg.MatchString(str) {
			r = append(r, reg)
		}
	}
	return r
}

func NewFinder(blacklist, whitelist []*regexp.Regexp) *Finder {
	return &Finder{
		blacklist: blacklist,
		whitelist: whitelist,
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

func (b *Finder) findBinary(path, filetype, section string, r io.Reader, rw ResultWriter) error {
	scanner := bufio.NewScanner(r)
	scanner.Split(split)
	for scanner.Scan() {
		t := scanner.Text()
		if t == "" {
			continue
		}
		if match, _ := matchRegexps(t, b.whitelist); match {
			continue
		}
		for _, keyword := range matchAllRegexps(t, b.blacklist) {
			rw.Write(path, filetype, section, keyword.String(), t)
		}
	}
	return nil
}

func (b *Finder) findElf(path string, r io.Reader, rw ResultWriter) error {
	f, err := elf.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, section := range f.Sections {
		// @see binutils strings.c
		// #define DATA_FLAGS (SEC_ALLOC | SEC_LOAD | SEC_HAS_CONTENTS)
		if (section.Flags & elf.SHF_ALLOC) != 0 {
			continue
		}
		if section.Type == elf.SHT_NOBITS {
			continue
		}
		
		src, err := section.Data()
		if err != nil {
			return err
		}
		err = b.findBinary(path, "elf", section.Name, bytes.NewReader(src), rw)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Finder) Find(path string, rw ResultWriter) error {
	r, err := os.Open(path)
	defer r.Close()
	if err != nil {
		return err
	}

	r1 := bytes.NewBuffer(nil)
	r2 := io.TeeReader(r, r1)
	if b.findElf(path, r1, rw) == nil {
		return nil
	}
	return b.findBinary(path, "bin", "", r2, rw)
}

func NewResultWriter(w io.Writer) ResultWriter {
	if w == nil {
		return &DummyResultWriter{}
	}
	return &ResultWriterImpl{
		w: w,
	}
}

func (rw *ResultWriterImpl) Write(path, filetype, section, keyword, text string) {
	r := regexp.MustCompile("[\t,\"]")
	t := r.ReplaceAllString(text, "")
	fmt.Fprintf(rw.w, "%s,%s,%s,%s,%s\n", path, filetype, section, keyword, t)
}

func (dw *DummyResultWriter) Write(path, filetype, section, keyword, text string) {
}

func check(format string, err error) {
	if err != nil {
		panic(fmt.Errorf(format, err))
	}
}

func readRegexps(filename string) ([]*regexp.Regexp, error) {
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
	return compileRegexps(regexps), nil
}

func buildSubstract(file string) (map[string]bool, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	substract := make(map[string]bool)
	for scanner.Scan() {
		substract[scanner.Text()] = true
	}
	return substract, nil
}

func eachFile(inputPath string, ignorePath []*regexp.Regexp, fn func(path string)) error {
	err := filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
		if match, _ := matchRegexps(inputPath, ignorePath); match {
			return nil
		}
		if ! info.Mode().IsRegular() {
			return err
		}
		fn(path)
		return nil
	})
	return err
}

func mainImplUsingGoroutine(blacklist, whitelist []*regexp.Regexp,
	inputpath string, ignorePath []*regexp.Regexp, rw ResultWriter) error {
	b := NewFinder(blacklist, whitelist)

	ch := make(chan string)
	wg := sync.WaitGroup{}
	worker := func() {
		defer wg.Done()
		for p := range ch {
			b.Find(p, rw)
		}
	}
	const sizeOfGorotine = 10
	wg.Add(sizeOfGorotine)
	for i:=0; i<sizeOfGorotine; i++ {
		go worker()
	}
	err := eachFile(inputpath, ignorePath, func(path string) {
		ch <- path
	})
	close(ch)
	wg.Wait()
	return err
}

func mainImplStanderd(blacklist, whitelist []*regexp.Regexp, inputpath string, ignorePath []*regexp.Regexp, rw ResultWriter) error {
	b := NewFinder(blacklist, whitelist)
	err := eachFile(inputpath, ignorePath, func(path string) {
		b.Find(path, rw)
	})
	return err
}

func printResult(newPassList io.Reader, result io.Writer, passList map[string]bool) error {
	scanner := bufio.NewScanner(newPassList)
	for scanner.Scan() {
		t := scanner.Text()
		if _, ok := passList[t]; ok {
			continue
		}
		fmt.Fprintln(result, t)
	}
	return nil
}

func main() {
	var (
		inputPath  = flag.String("i", "", "input path")
		ignorePathFile = flag.String("ignore", "", "ignore path file")
		passListFile = flag.String("pass", "", "pass list file")
		blackListFile = flag.String("black", "", "regexp file(blacklist)")
		whiteListFile = flag.String("white", "", "regexp file(whitelist)")
		newPathList = flag.String("new_pass_list", "-", "new pass list")
		result = flag.String("result", "-", "result(new_pass_list - pass_list)")
		err error
	)
	flag.Parse()

	var ignorePath []*regexp.Regexp
	if *ignorePathFile != "" {
		ignorePath, err = readRegexps(*ignorePathFile)
		check("-ignore error: %v", err)
	}

	blacklist, err := readRegexps(*blackListFile)
	if err != nil {
		check("-black error: %v", err)
	}
	whitelist, err := readRegexps(*whiteListFile)
	check("-white error: %v", err)

	fp := os.Stdout
	if *newPathList != "-" {
		fp, err = os.Create(*newPathList)
		check("-new_pass_list error: %v", err)
		defer fp.Close()
	}
	buffer := bytes.NewBuffer(nil)
	
	npl := NewResultWriter(io.MultiWriter(fp, buffer))
	err = mainImplUsingGoroutine(blacklist, whitelist, *inputPath, ignorePath, npl)
	check("finding err: %v", err)

	rslt := os.Stdout
	if *result != "-" {
		rslt, err = os.Create(*result)
		check("-result error: %v", err)
		defer rslt.Close()
	}

	var passList map[string]bool
	if *passListFile != "" {
		passList, err = buildSubstract(*passListFile)
		check("-pass error: %v", err)
	}

	err = printResult(buffer, rslt, passList)
	check("result error: %v", err)
}
