package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"text/scanner"
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

func IsVoid(types []string) bool {
	result := false
	for _, typ := range types {
		if typ == "*" {
			return false
		}
		if typ == "void" {
			result = true
		}
	}
	return result
}

func (a Arg) String() string {
	var sb strings.Builder
	for _, typ := range a.typ {
		sb.WriteString(typ)
		sb.WriteString(" ")
	}
	sb.WriteString(a.name)
	return sb.String()
}

func (a Arg) WriteExpectMock(w io.Writer) {
	fmt.Fprintf(w, ".withParameter(\"%s\", %s)", a.name, a.name)
}

func (a Arg) WriteActualMock(w io.Writer) {
	fmt.Fprintf(w, ".withParameter(\"%s\", %s)", a.name, a.name)
}

func (fd FunctionDeclaration) WriteExpectFunction(w io.Writer) {
	bw := bufio.NewWriter(w)
	bw.WriteString("void expect_")
	bw.WriteString(fd.name)
	bw.WriteString("(")
	for i, arg := range fd.args {
		if i != 0 {
			bw.WriteString(", ")
		}
		bw.WriteString(arg.String())
	}
	if ! IsVoid(fd.typ) {
		bw.WriteString(", ")
		bw.WriteString(strings.Join(fd.typ, " "))
		bw.WriteString(" retval")
	}
	bw.WriteString(") {\n")
	bw.WriteString("{\n")

	fmt.Fprintf(bw, "    mock().expectOneCall(\"%s\")", fd.name)
	for _, arg := range fd.args {
		bw.WriteString("\n          ")
		arg.WriteExpectMock(bw)
	}
	if ! IsVoid(fd.typ) {
		bw.WriteString("\n          ")
		bw.WriteString(".andReturnValue(retval)")
	}
	bw.WriteString(";\n")

	bw.WriteString("}\n")

	bw.Flush()
}

func (fd FunctionDeclaration) WriteActualFunction(w io.Writer) {
	bw := bufio.NewWriter(w)
	bw.WriteString(strings.Join(fd.typ, " "))
	bw.WriteString(" ")
	bw.WriteString(fd.name)
	bw.WriteString("(")
	for i, arg := range fd.args {
		if i != 0 {
			bw.WriteString(", ")
		}
		bw.WriteString(arg.String())
	}
	bw.WriteString(") {\n")
	bw.WriteString("{\n")

	fmt.Fprintf(bw, "    mock().actualOneCall(\"%s\")", fd.name)
	for _, arg := range fd.args {
		bw.WriteString("\n          ")
		arg.WriteActualMock(bw)

	}
	if ! IsVoid(fd.typ) {
		bw.WriteString("\n          ")
		bw.WriteString(".returnValue()")
	}
	bw.WriteString(";\n")

	bw.WriteString("}\n")


	bw.Flush()
}

func main() {
	l := new(Lexer)
	l.Init(strings.NewReader(os.Args[1]))
	yyParse(l)
	fd := l.result.(FunctionDeclaration)

	fd.WriteExpectFunction(os.Stdout)
	fmt.Fprintf(os.Stdout, "\n")
	fd.WriteActualFunction(os.Stdout)
}
