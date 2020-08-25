package main

import(
	"testing"
)

func TestExample(t *testing.T) {
	i := add(1, 2)
	if (i != 3) {
		t.Fatalf("expected = %d, actual = %d\n", 3, i)
	}
}
