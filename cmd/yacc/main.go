package main

import (
	"fmt"
	"text/scanner"
	"os"
	"strings"
)

type Lexer struct {
	scanner.Scanner
	result Expression
}

const debug = true

func debugPrintf(format string, a ...interface{}) (n int, err error) {
	if debug != true {
		return 0, nil
	}
	return fmt.Printf(format, a...)
}

func (l *Lexer) Lex(lval *yySymType) int {
	token := int(l.Scan())
	debugPrintf("Scan() returns %d,", token)
	switch token {
	case scanner.Int:
		token = NUMBER
		debugPrintf("Lex() returns NUMBER,")
	case scanner.Ident:
		token = IDENT
		debugPrintf("Lex() returns IDENT,")
	default:
		debugPrintf("Lex() returns other(perhaps ascii or EOF),")
	}
	lval.token = Token{literal: l.TokenText()}
	debugPrintf("literal = %s\n", lval.token.literal)
	return token
}

func (l *Lexer) Error(e string) {
	panic(e)
}

func main() {
	l := new(Lexer)
	l.Init(strings.NewReader(os.Args[1]))
	yyParse(l)
	fmt.Printf("%#v\n", l.result)

	fmt.Printf("%s", os.Args[1])
	fmt.Printf(`
{
}`)

// 	r := l.result
// 	switch r.(type) {
// 	case BinOpExpr:
// 		left := r.(BinOpExpr).left
// 		right := r.(BinOpExpr).right
// 		fmt.Printf("%#v\n", left.(NumExpr).literal)
// 		fmt.Printf("%#v\n", right)
// 	}
}
