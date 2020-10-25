package main

import(
	"testing"
)

func BenchmarkOneFile(b *testing.B) {
	blacklist := []string{"(?i)hellO", "(?i)wOr"}
	whitelist := []string{"WOR"}
	b.ResetTimer()
	err := mainImpl(blacklist, whitelist, ".")
	if err != nil {
		b.Fatalf("unexpected err = %W\n", err)
	}
}

