package main

import(
	"testing"
)

var blacklist = []string{"(?i)hellO", "(?i)wOr"}
var whitelist = []string{"WOR"}

func BenchmarkOneFile(b *testing.B) {
	b.ResetTimer()
	err := mainImplStanderd(blacklist, whitelist, ".")
	if err != nil {
		b.Fatalf("unexpected err = %W\n", err)
	}
}

func BenchmarkParentDirectory(b *testing.B) {
	b.ResetTimer()
	err := mainImplStanderd(blacklist, whitelist, "../..")
	if err != nil {
		b.Fatalf("unexpected err = %W\n", err)
	}
}

func BenchmarkStanderd(b *testing.B) {
	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		err := mainImplStanderd(blacklist, whitelist, "../..")
		if err != nil {
			b.Fatalf("unexpected err = %W\n", err)
		}
	}
}

func BenchmarkUsingGoroutine(b *testing.B) {
	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		err := mainImplUsingGoroutine(blacklist, whitelist, "../..")
		if err != nil {
			b.Fatalf("unexpected err = %W\n", err)
		}
	}
}

