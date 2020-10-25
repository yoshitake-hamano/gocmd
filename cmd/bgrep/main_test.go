package main

import(
	"testing"
)

var blacklist = []string{"(?i)hellO", "(?i)wOr"}
var whitelist = []string{"WOR"}

func BenchmarkOneFile(b *testing.B) {
	b.ResetTimer()
	err := mainImpl(blacklist, whitelist, ".")
	if err != nil {
		b.Fatalf("unexpected err = %W\n", err)
	}
}

func BenchmarkParentDirectory(b *testing.B) {
	b.ResetTimer()
	err := mainImpl(blacklist, whitelist, "../..")
	if err != nil {
		b.Fatalf("unexpected err = %W\n", err)
	}
}

