#!/bin/sh

BIN_DIR=../bin
HELLO=${BIN_DIR}/hello
CREATEMOCK=${BIN_DIR}/createmock

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

function diffString() {
    expect="${1}"
    actual="${2}"

    expectFile=`mktemp`
    echo "${expect}" > ${expectFile}
    actualFile=`mktemp`
    echo "${actual}" > ${actualFile}
    echo "-: expect, +: actual"
    diff -u ${expectFile} ${actualFile}
    rm ${expectFile} ${actualFile}
}

function assertStringEquals() {
    expect="${1}"
    actual="${2}"
    if [ "${expect}" != "${actual}" ]; then
        echo "assertStringEquals fails"
        echo "expect = ${expect}, actual = ${actual}"

        diffString "${expect}" "${actual}"
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

function assertReturnsOK() {
    s=`${1} "${2}"`
    result=${?}
    echo "${1}"
    echo "${s}"
    assertIntEquals 0 ${result}
}

function assertExec() {
    s=`${1} "${2}"`
    result=${?}
    echo "${1} ${2}"
    echo "${s}"
    assertIntEquals ${3} ${result}
    assertStringEquals "${4}" "${s}"
}

assertStringEquals "abc" "abc"

function testSuiteHello() {
    s=`${HELLO}`
    assertIntEquals 0 ${?}
    assertStringEquals "Hello" "${s}"
}

function testSuiteCreatemock() {
    assertReturnsOK ${CREATEMOCK} "void sum(void)"
    assertReturnsOK ${CREATEMOCK} "void sum(int a)"
    assertReturnsOK ${CREATEMOCK} "void sum(int)"
    assertReturnsOK ${CREATEMOCK} "unsigned int sum(int a, int b)"
    assertReturnsOK ${CREATEMOCK} "unsigned int sum(int a, int *b)"

    assertExec ${CREATEMOCK} "void sum(int a)" 0 'void expect_sum(int a)
{
    mock().expectOneCall("sum")
          .withParameter("a", a);
}

void sum(int a)
{
    mock().actualOneCall("sum")
          .withParameter("a", a);
}'
}

testSuiteHello
testSuiteCreatemock
