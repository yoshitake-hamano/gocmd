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
%type<expr> expr
%type<expr> arg
%token<token> NUMBER IDENT

%left ','

%%

top
    : expr '(' arg ')'
    {
        expr := $1.([]string)
        typ  := expr[0:len(expr)-1]
        name := expr[len(expr)-1]

        $$ = FunctionDeclaration{typ: typ, name: name, args: $3.([]Arg)}
        yylex.(*Lexer).result = $$
    }

arg
    : expr
    {
        expr := $1.([]string)
        typ  := expr[0:len(expr)-1]
        name := expr[len(expr)-1]

        arg := Arg{typ: typ, name: name}
        $$ = []Arg{arg}
    }
    | arg ',' arg
    {
        lhd := $1.([]Arg)
        rhd := $3.([]Arg)
        $$ = append(lhd, rhd...)
    }

expr
    : NUMBER
    {
        expr := string($1.literal)
        $$ = []string{expr}
    }
    | IDENT
    {
        expr := string($1.literal)
        $$ = []string{expr}
    }
    | expr expr
    {
        lhd := $1.([]string)
        rhd := $2.([]string)
        $$ = append(lhd, rhd...)
    }

%%
