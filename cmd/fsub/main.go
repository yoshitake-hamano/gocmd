package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
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

func main() {
	var (
		srcfile  = flag.String("src", "", "src file")
		substractfile  = flag.String("sub", "", "substract file")
	)
	flag.Parse()

	sub, err := buildSubstract(*substractfile)
	check(err)

	f, err := os.Open(*srcfile)
	check(err)
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := scanner.Text()
		if _, ok := sub[t]; ok {
			continue
		}
		fmt.Println(t)
	}
}
