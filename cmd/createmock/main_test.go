package main

import(
	"testing"
)

func TestExample(t *testing.T) {
	var tests = []struct {
		typ    []string
		expect CpputestType
	}{
		{typ: []string{"void"}, expect: typeVoid},
		{typ: []string{"int"},  expect: typeInt},
	}
	for _, test := range tests {
		output := getCpputestType(test.typ)
		if (output != test.expect) {
			t.Fatalf("expected = %d, actual = %d\n", test.expect, output)
		}
	}
}
