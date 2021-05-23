package main

import (
	"fmt"
)

type stringsFlag []string

func (s *stringsFlag) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringsFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}
