package main

import (
	"bufio"
	"debug/elf"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type Blackouter struct {
	regexps []*regexp.Regexp
}

func NewBlackouter(searchWords []string) *Blackouter {
	regexps := make([]*regexp.Regexp, 0, len(searchWords))
	for _, word := range searchWords {
		r := regexp.MustCompile(word)
		regexps = append(regexps, r)
	}
	return &Blackouter{
		regexps: regexps,
	}
}

func (b *Blackouter) blackout(src []byte, fn func([]byte)) []byte {
	dest := append(src[:0:0], src...)
	for _, r := range b.regexps {
		if fn != nil {
			for _, match := range r.FindAll(dest, len(dest)) {
				fn(match)
			}
		}
		dest = r.ReplaceAllFunc(dest, func(repl []byte) []byte {
			return []byte(strings.Repeat("*", len(repl)))
		})
	}
	return dest
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func readConfig(filename string) ([]string, error) {
	config := make([]string, 0)
	fp, err := os.Open(filename)
	defer fp.Close()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		config = append(config, scanner.Text())
	}
	return config, nil
}

type stringsFlag []string

func (s *stringsFlag) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringsFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

var sections stringsFlag

func main() {
	var (
		inputfile  = flag.String("i", "", "input file")
		outputfile = flag.String("o", "", "output file")
		regexpfile = flag.String("r", "", "regexp file")
	)
	flag.Var(&sections, "s", "sections to black out")
	flag.Parse()
	for _, s := range sections {
		fmt.Printf("%s\n", s)
	}

	filedata, err := ioutil.ReadFile(*inputfile)
	check(err)

	config, err := readConfig(*regexpfile)
	check(err)
	b := NewBlackouter(config)
	f, err := elf.Open(*inputfile)
	check(err)
	defer f.Close()

	for _, section := range f.Sections {
		if section.Name != ".rodata" {
			continue
		}

		src, err := section.Data()
		check(err)
		dest := b.blackout(src, func(match []byte) {
			fmt.Printf("%s: %s\n", *inputfile, string(match))
		})
		if int(section.Size) != len(dest) {
			check(fmt.Errorf("mismatch regexp size %s(before %d, after %d)",
				section.Name, section.Size, len(dest)))
		}

		sectionStart := section.Offset
		sectionEnd := section.Offset + section.Size
		tmp := append(filedata[0:sectionStart], dest...)
		filedata = append(tmp, filedata[sectionEnd:]...)
	}

	err = ioutil.WriteFile(*outputfile, filedata, 0644)
	check(err)

	// d, err := f.DWARF()
	// check(err)
	//
	// r := d.Reader()
	// for {
	// 	e, err := r.Next()
	// 	if e == nil || err != nil {
	// 		fmt.Print(err)
	// 		break
	// 	}
	// 	fmt.Printf("Tag %v\n", e.Tag)
	// 	for i, f := range e.Field {
	// 		fmt.Printf("field[%d] %v, %v, %v\n", i, f.Class, f.Attr, f.Val)
	// 	}
	// }

	// fmt.Print(d)
}
