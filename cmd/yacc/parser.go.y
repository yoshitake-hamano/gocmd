%{

// C language BNF
// https://cs.wmich.edu/~gupta/teaching/cs4850/sumII06/The%20syntax%20of%20C%20in%20Backus-Naur%20form.htm
package main

type Expression interface{}
type Token struct {
    token   int
    literal string
}

type NumExpr struct {
    literal string
}
type IdentExpr struct {
    literal string
}
%}

%union{
    token Token
    expr  Expression
}

%type<expr> top
%type<expr> expr
%token<token> NUMBER IDENT

%left ','

%%

top
    : expr
    {
        $$ = $1
        yylex.(*Lexer).result = $$
    }

expr
    : NUMBER
    {
        $$ = NumExpr{literal: $1.literal}
    }
    | IDENT
    {
        $$ = IdentExpr{literal: $1.literal}
    }

%%
