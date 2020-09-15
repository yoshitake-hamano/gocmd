%{

// C language BNF
// https://cs.wmich.edu/~gupta/teaching/cs4850/sumII06/The%20syntax%20of%20C%20in%20Backus-Naur%20form.htm
package main

type Token struct {
    token   int
    literal string
}

type Expression interface{}
type NumExpr struct {
    literal string
}
type IdentExpr struct {
    literal string
}
type TypeExpr struct {
    literal string
}
type Exprs struct {
    exprs []Expression
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
        lhd := $1.(Args)
        rhd := $3.(Args)
        $$ = Args{args: append(lhd.args, rhd.args...)}
    }

expr
    : NUMBER
    {
        expr := NumExpr{literal: $1.literal}
        $$ = Exprs{exprs: []Expression{expr}}
    }
    | IDENT
    {
        debugPrintf("expr %v\n", $1)
        expr := IdentExpr{literal: $1.literal}
        $$ = Exprs{exprs: []Expression{expr}}
    }
    | expr expr
    {
        debugPrintf("expr expr %v %v\n", $1, $2)

        lhd := $1.(Exprs)
        rhd := $2.(Exprs)
        $$ = Exprs{exprs: append(lhd.exprs, rhd.exprs...)}
    }

%%
