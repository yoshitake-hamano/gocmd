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
type TypeExpr struct {
    literal string
}
type PairExprs struct {
    lhd Expression
    rhd Expression
}
type Arg struct {
    arg Expression
}
type PairArgs struct {
    lhd Expression
    rhd Expression
}
type FunctionDeclaration struct {
    typ  Expression
    name Expression
    args Expression
}
%}

%union{
    token Token
    expr  Expression
}

%type<expr> top
%type<expr> typ
%type<expr> name
%type<expr> expr
%type<expr> arg
%token<token> NUMBER IDENT

%left ','

%%

top
    : expr '(' arg ')'
    {
        $$ = FunctionDeclaration{typ: $1, args: $3}
        yylex.(*Lexer).result = $$
    }

typ
    : expr
    {
        $$ = $1
    }

name
    : IDENT
    {
        $$ = IdentExpr{literal: $1.literal}
    }

arg
    : expr
    {
        $$ = Arg{arg: $1}
    }
    | arg ',' arg
    {
        $$ = PairArgs{lhd: $1, rhd: $3}
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
    | expr expr
    {
        $$ = PairExprs{lhd: $1, rhd: $2}
    }

%%
