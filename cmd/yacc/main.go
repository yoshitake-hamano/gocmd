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

const debug = false

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
	case '*':
		token = IDENT
		debugPrintf("Lex() returns *,")
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
	fd := l.result.(FunctionDeclaration)


	fmt.Printf("%s\n", os.Args[1])
	fmt.Printf("\n")


	fmt.Printf("  fd name = %s\n", fd.name)
	for _, typ := range fd.typ {
		fmt.Printf("    fd type : %s\n", typ)
	}
	for _, arg := range fd.args {
		fmt.Printf("      arg name = %s\n", arg.name)
		for _, typ := range arg.typ {
			fmt.Printf("        arg type : %s\n", typ)
		}
	}
}
