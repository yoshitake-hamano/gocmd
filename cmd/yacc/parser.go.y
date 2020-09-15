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
type Args struct {
    args []Arg
}
type FunctionDeclaration struct {
    typ  Expression
    args Expression
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
        $$ = FunctionDeclaration{typ: $1, args: $3}
        yylex.(*Lexer).result = $$
    }

arg
    : expr
    {
        arg := Arg{arg: $1}
        $$ = Args{args: []Arg{arg}}
    }
    | arg ',' arg
    {
        debugPrintf("arg,arg %v,%v\n", $1, $3)
        lhd := $1.(Args)
        rhd := $3.(Args)
        $$ = Args{args: append(lhd.args, rhd.args...)}
    }

expr
    : NUMBER
    {
        $$ = NumExpr{literal: $1.literal}
    }
    | IDENT
    {
        debugPrintf("expr %v\n", $1)
        $$ = IdentExpr{literal: $1.literal}
    }
    | expr expr
    {
        debugPrintf("expr expr %v %v\n", $1, $2)
        $$ = PairExprs{lhd: $1, rhd: $2}
    }

%%
