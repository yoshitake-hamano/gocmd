%{

// C language BNF
// https://cs.wmich.edu/~gupta/teaching/cs4850/sumII06/The%20syntax%20of%20C%20in%20Backus-Naur%20form.htm
package main

type Token struct {
    token   int
    literal string
}

type Expression interface{}
type Arg struct {
    typ  []string
    name string
}
type FunctionDeclaration struct {
    typ  []string
    name string
    args []Arg
}
%}

%union{
    token Token
    expr  Expression
}

%type<expr> top
%type<expr> fnctn
%type<expr> expr
%type<expr> arg
%token<token> NUMBER IDENT EOF

%left ','

%%

top
    : fnctn
    {
        debugPrintf("Syntax: top = fnctn(%v)\n", $1)
        fd := $1.(FunctionDeclaration)
        $$ = []FunctionDeclaration{fd}
        yylex.(*Lexer).result = $$
    }
    | top ';' top
    {
        debugPrintf("Syntax: top = top(%v) ; top(%v)\n", $1, $3)
        lhd := $1.([]FunctionDeclaration)
        rhd := $3.([]FunctionDeclaration)
        $$ = append(lhd, rhd...)
        yylex.(*Lexer).result = $$
    }
    | top ';' EOF
    {
        debugPrintf("Syntax: top = top(%v) ; EOF\n", $1)
        $$ = $1.([]FunctionDeclaration)
        yylex.(*Lexer).result = $$
    }

fnctn
    : expr '(' arg ')'
    {
        debugPrintf("Syntax: fnctn = expr(%v) ( arg(%v) )\n", $1, $3)
        expr := $1.([]string)
        typ  := expr[0:len(expr)-1]
        name := expr[len(expr)-1]

        $$ = FunctionDeclaration{typ: typ, name: name, args: $3.([]Arg)}
    }

arg
    : expr
    {
        debugPrintf("Syntax: arg = expr(%v)\n", $1)
        expr := $1.([]string)
        typ  := expr[0:len(expr)-1]
        name := expr[len(expr)-1]

        arg := Arg{typ: typ, name: name}
        $$ = []Arg{arg}
    }
    | arg ',' arg
    {
        debugPrintf("Syntax: arg = arg(%v), arg(%v)\n", $1, $3)
        lhd := $1.([]Arg)
        rhd := $3.([]Arg)
        $$ = append(lhd, rhd...)
    }

expr
    : NUMBER
    {
        debugPrintf("Syntax: expr = NUMBER(%v)\n", $1)
        expr := string($1.literal)
        $$ = []string{expr}
    }
    | IDENT
    {
        debugPrintf("Syntax: expr = IDENT(%v)\n", $1)
        expr := string($1.literal)
        $$ = []string{expr}
    }
    | expr expr
    {
        debugPrintf("Syntax: expr = expr(%v) expr(%v)\n", $1, $2)
        lhd := $1.([]string)
        rhd := $2.([]string)
        $$ = append(lhd, rhd...)
    }

%%
