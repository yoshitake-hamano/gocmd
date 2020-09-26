package main

import (
	"bufio"
	"flag"
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
	debugPrintf("Lex:    Scan() returns %d, ", token)
	switch token {
	case scanner.Int:
		token = NUMBER
		debugPrintf("Lex() returns NUMBER, ")
	case scanner.Ident:
		token = IDENT
		debugPrintf("Lex() returns IDENT, ")
	case scanner.EOF:
		token = EOF
		debugPrintf("Lex() returns EOF, ")
	case '*':
		token = IDENT
		debugPrintf("Lex() returns *, ")
	default:
		debugPrintf("Lex() returns other(perhaps ascii), ")
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
	bw.WriteString(")\n")
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
	bw.WriteString(")\n")
	bw.WriteString("{\n")

	if ! IsVoid(fd.typ) {
		fmt.Fprintf(bw, "    return mock().actualCall(\"%s\")", fd.name)
	} else {
		fmt.Fprintf(bw, "    mock().actualCall(\"%s\")", fd.name)
	}
	for _, arg := range fd.args {
		bw.WriteString("\n          ")
		arg.WriteActualMock(bw)

	}
	if ! IsVoid(fd.typ) {
		bw.WriteString("\n          ")
		bw.WriteString(".returnIntValue()")
	}
	bw.WriteString(";\n")

	bw.WriteString("}\n")


	bw.Flush()
}

func main() {
	var (
		file = flag.String("file", "", "the c header file")
		arg  = flag.String("arg",  "", "the c function declaration")
		r io.Reader
	)
	flag.Parse()

	l := new(Lexer)
	if *file != "" {
		var err error
		r, err = os.Open(*file)
		if err != nil {
			panic(err)
		}
	}
	if *arg != "" {
		r = strings.NewReader(*arg)
	}
	l.Init(r)
	yyParse(l)
	fds := l.result.([]FunctionDeclaration)

	for _, fd := range fds {
		fd.WriteExpectFunction(os.Stdout)
		fmt.Fprintf(os.Stdout, "\n")
		fd.WriteActualFunction(os.Stdout)
	}
}
