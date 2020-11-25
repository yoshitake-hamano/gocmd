package main

import(
	"os"
	"testing"
)

var blacklist = compileRegexps([]string{"(?i)hellO", "(?i)wOr"})
var whitelist = compileRegexps([]string{"WOR"})
var ignorePath = compileRegexps([]string{"xxx"})

func TestOneFile(t *testing.T) {
	rw := NewResultWriter(os.Stdout)
	err := mainImplStanderd(blacklist, whitelist, ".", ignorePath, rw)
	if err != nil {
		t.Fatalf("unexpected err = %W\n", err)
	}
}

func TestParentDirectory(t *testing.T) {
	rw := NewResultWriter(os.Stdout)
	err := mainImplStanderd(blacklist, whitelist, "../..", ignorePath, rw)
	if err != nil {
		t.Fatalf("unexpected err = %W\n", err)
	}
}

func BenchmarkStanderd(b *testing.B) {
	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		rw := NewResultWriter(os.Stdout)
		err := mainImplStanderd(blacklist, whitelist, "../..", ignorePath, rw)
		if err != nil {
			b.Fatalf("unexpected err = %v\n", err)
		}
	}
}

func BenchmarkUsingGoroutine(b *testing.B) {
	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		rw := NewResultWriter(os.Stdout)
		err := mainImplUsingGoroutine(blacklist, whitelist, "../..", ignorePath, rw)
		if err != nil {
			b.Fatalf("unexpected err = %v\n", err)
		}
	}
}

