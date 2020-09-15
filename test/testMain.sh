#!/bin/sh

BIN_DIR=../bin
HELLO=${BIN_DIR}/hello
YACC=${BIN_DIR}/yacc

function dumpStack() {
    local i=0
    local line_no
    local function_name
    local file_name
    while caller $i ;do ((i++)) ;done | while read line_no function_name file_name;do echo "\t$file_name:$line_no\t$function_name" ;done >&2
}

function flunk() {
    dumpStack
    exit 1
}

function assertStringEquals() {
    expect="${1}"
    actual="${2}"
    if [ "${expect}" != "${actual}" ]; then
        echo "assertStringEquals fails"
        echo "expect = ${expect}, actual = ${actual}"
        flunk
    fi
}

function assertIntEquals() {
    expect=${1}
    actual=${2}
    if [ ! ${expect} -eq ${actual} ]; then
        echo "assertIntEquals fails"
        echo "expect = ${expect}, actual = ${actual}"
        flunk
    fi
}

assertStringEquals "abc" "abc"

function testHello() {
    s=`${HELLO}`
    assertIntEquals 0 ${?}
    assertStringEquals "Hello" "${s}"
}

function testYacc() {
    # s=`${YACC} "void sum(int a, int b)"`
    s=`${YACC} "void sum int a, int b"`
    result=${?}
    echo "${s}"
    assertIntEquals 0 ${result}
    assertStringEquals "main.BinOpExpr{left:main.NumExpr{literal:\"1\"}, operator:43, right:main.NumExpr{literal:\"2\"}}" "${s}"
}

testHello
testYacc
