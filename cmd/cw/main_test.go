package main

import(
	"testing"
)

var blacklist = []string{"(?i)hellO", "(?i)wOr"}
var whitelist = []string{"WOR"}

func TestOneFile(t *testing.T) {
	err := mainImplStanderd(blacklist, whitelist, ".")
	if err != nil {
		t.Fatalf("unexpected err = %W\n", err)
	}
}

func TestParentDirectory(t *testing.T) {
	err := mainImplStanderd(blacklist, whitelist, "../..")
	if err != nil {
		t.Fatalf("unexpected err = %W\n", err)
	}
}

func BenchmarkStanderd(b *testing.B) {
	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		*silent = true
		err := mainImplStanderd(blacklist, whitelist, "../../..")
		if err != nil {
			b.Fatalf("unexpected err = %v\n", err)
		}
	}
}

func BenchmarkUsingGoroutine(b *testing.B) {
	b.ResetTimer()
	for i:=0; i<b.N; i++ {
		*silent = true
		err := mainImplUsingGoroutine(blacklist, whitelist, "../../..")
		if err != nil {
			b.Fatalf("unexpected err = %v\n", err)
		}
	}
}

