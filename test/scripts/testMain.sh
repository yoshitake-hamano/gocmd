#!/bin/sh

BIN_DIR=../../bin
HELLO=${BIN_DIR}/hello
CREATEMOCK="${BIN_DIR}/createmock -arg"
EXCELGO="${BIN_DIR}/excelgo"
EXCELGANTT="${BIN_DIR}/excelgantt"

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

function assertIntNotEquals() {
    expect=${1}
    actual=${2}
    if [ ${expect} -eq ${actual} ]; then
        echo "assertIntEquals fails"
        echo "expect = ${expect}, actual = ${actual}"
        flunk
    fi
}

function assertReturnsOK() {
    s=`${1} "${2}"`
    result=${?}
    echo "${1} ${2}"
    echo "${s}"
    assertIntEquals 0 ${result}
}

function assertReturnsOKMultiArgs() {
    s=`${1} ${2}`
    result=${?}
    echo "${1} ${2}"
    echo "${s}"
    assertIntEquals 0 ${result}
}

function assertReturnsNGMultiArgs() {
    s=`${1} ${2}`
    result=${?}
    echo "${1} ${2}"
    echo "${s}"
    assertIntNotEquals 0 ${result}
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

function testSuiteCreatemockReturnsOK() {
    assertReturnsOK "${CREATEMOCK}" "void sum(void)"
    assertReturnsOK "${CREATEMOCK}" "void sum(int a)"
    assertReturnsOK "${CREATEMOCK}" "void sum(int)"
    assertReturnsOK "${CREATEMOCK}" "unsigned int sum(int a, int b)"
    assertReturnsOK "${CREATEMOCK}" "unsigned int sum(int a, int *b)"
}

function testSuiteCreatemockGlobalVariable() {
    assertExec "${CREATEMOCK}" "int value" 0 ''
    assertExec "${CREATEMOCK}" "int value;" 0 ''
}

function testSuiteCreatemockSimpleFunction() {
    assertExec "${CREATEMOCK}" "void sum(int a)" 0 'void expect_sum(int a)
{
    mock().expectOneCall("sum")
          .withParameter("a", a);
}

void sum(int a)
{
    mock().actualCall("sum")
          .withParameter("a", a);
}'
    assertExec "${CREATEMOCK}" "void sum(int a);" 0 'void expect_sum(int a)
{
    mock().expectOneCall("sum")
          .withParameter("a", a);
}

void sum(int a)
{
    mock().actualCall("sum")
          .withParameter("a", a);
}'
}

function testSuiteCreatemockReturnType() {
    assertExec "${CREATEMOCK}" "int sum(int a)" 0 'void expect_sum(int a, int retval)
{
    mock().expectOneCall("sum")
          .withParameter("a", a)
          .andReturnValue(retval);
}

int sum(int a)
{
    return mock().actualCall("sum")
          .withParameter("a", a)
          .returnIntValue();
}'

    assertExec "${CREATEMOCK}" "unsigned int sum(int a)" 0 'void expect_sum(int a, unsigned int retval)
{
    mock().expectOneCall("sum")
          .withParameter("a", a)
          .andReturnValue(retval);
}

unsigned int sum(int a)
{
    return mock().actualCall("sum")
          .withParameter("a", a)
          .returnUnsignedIntValue();
}'

    assertExec "${CREATEMOCK}" "long sum(int a)" 0 'void expect_sum(int a, long retval)
{
    mock().expectOneCall("sum")
          .withParameter("a", a)
          .andReturnValue(retval);
}

long sum(int a)
{
    return mock().actualCall("sum")
          .withParameter("a", a)
          .returnLongIntValue();
}'

    assertExec "${CREATEMOCK}" "unsigned long sum(int a)" 0 'void expect_sum(int a, unsigned long retval)
{
    mock().expectOneCall("sum")
          .withParameter("a", a)
          .andReturnValue(retval);
}

unsigned long sum(int a)
{
    return mock().actualCall("sum")
          .withParameter("a", a)
          .returnUnsignedLongIntValue();
}'

    assertExec "${CREATEMOCK}" "long long sum(int a)" 0 'void expect_sum(int a, long long retval)
{
    mock().expectOneCall("sum")
          .withParameter("a", a)
          .andReturnValue(retval);
}

long long sum(int a)
{
    return mock().actualCall("sum")
          .withParameter("a", a)
          .returnLongLongIntValue();
}'

    assertExec "${CREATEMOCK}" "unsigned long long sum(int a)" 0 'void expect_sum(int a, unsigned long long retval)
{
    mock().expectOneCall("sum")
          .withParameter("a", a)
          .andReturnValue(retval);
}

unsigned long long sum(int a)
{
    return mock().actualCall("sum")
          .withParameter("a", a)
          .returnUnsignedLongLongIntValue();
}'

    assertExec "${CREATEMOCK}" "double sum(int a)" 0 'void expect_sum(int a, double retval)
{
    mock().expectOneCall("sum")
          .withParameter("a", a)
          .andReturnValue(retval);
}

double sum(int a)
{
    return mock().actualCall("sum")
          .withParameter("a", a)
          .returnDoubleValue();
}'
}

function testSuiteCreatemockArgumentType() {
    assertExec "${CREATEMOCK}" "int sum(unsigned int a)" 0 'void expect_sum(unsigned int a, int retval)
{
    mock().expectOneCall("sum")
          .withParameter("a", a)
          .andReturnValue(retval);
}

int sum(unsigned int a)
{
    return mock().actualCall("sum")
          .withParameter("a", a)
          .returnIntValue();
}'

    assertExec "${CREATEMOCK}" "int sum(long a)" 0 'void expect_sum(long a, int retval)
{
    mock().expectOneCall("sum")
          .withParameter("a", a)
          .andReturnValue(retval);
}

int sum(long a)
{
    return mock().actualCall("sum")
          .withParameter("a", a)
          .returnIntValue();
}'

    assertExec "${CREATEMOCK}" "int sum(double a)" 0 'void expect_sum(double a, int retval)
{
    mock().expectOneCall("sum")
          .withParameter("a", a)
          .andReturnValue(retval);
}

int sum(double a)
{
    return mock().actualCall("sum")
          .withParameter("a", a)
          .returnIntValue();
}'
    assertExec "${CREATEMOCK}" "int sum(char *s)" 0 'void expect_sum(char * s, int retval)
{
    mock().expectOneCall("sum")
          .withParameter("s", s)
          .andReturnValue(retval);
}

int sum(char * s)
{
    return mock().actualCall("sum")
          .withParameter("s", s)
          .returnIntValue();
}'

    assertExec "${CREATEMOCK}" "int sum(void *p, int p_size)" 0 'void expect_sum(void * p, int p_size, int retval)
{
    mock().expectOneCall("sum")
          // case1: if compare address
          // .withParameter("p", p)
          // case2: if compare value of address
          .withMemoryBufferParameter("p", (const unsigned char *)p, p_size)
          // case3: if output value
          // .withOutputParameterReturning("p", (const void *)p, p_size)
          .withParameter("p_size", p_size)
          .andReturnValue(retval);
}

int sum(void * p, int p_size)
{
    return mock().actualCall("sum")
          // case1: if compare address
          // .withParameter("p", p)
          // case2: if compare value of address
          .withMemoryBufferParameter("p", (const unsigned char *)p, p_size)
          // case3: if output value
          // .withOutputParameter("p", (void *)p)
          .withParameter("p_size", p_size)
          .returnIntValue();
}'
}

function testSuiteCreatemock() {
    testSuiteCreatemockReturnsOK
    testSuiteCreatemockGlobalVariable
    testSuiteCreatemockSimpleFunction
    testSuiteCreatemockReturnType
    testSuiteCreatemockArgumentType
}

function testSuiteExcelgoReturnsOK() {
    assertReturnsOKMultiArgs "${EXCELGO}" "-v -excel ../excelgo/variable.xlsx -sheet Sheet1 -target TARGET1 -template ../excelgo/template -output ../build"
}

function testSuiteExcelgoReturnsNG() {
    assertReturnsNGMultiArgs "${EXCELGO}" "-v -excel nofile.xlsx -sheet Sheet1 -target TARGET1 -template ../excelgo/template -output ../build"
    assertReturnsNGMultiArgs "${EXCELGO}" "-v -excel ../excelgo/variable.xlsx -sheet NoSheet -target TARGET1 -template ../excelgo/template -output ../build"
    assertReturnsNGMultiArgs "${EXCELGO}" "-v -excel ../excelgo/variable.xlsx -sheet Sheet1 -target NOTARGET -template ../excelgo/template -output ../build"
}

function testSuiteExcelgo() {
    testSuiteExcelgoReturnsOK
    testSuiteExcelgoReturnsNG
}

function testSuiteExcelganttReturnsOK() {
    assertReturnsOKMultiArgs "${EXCELGANTT}" "-v -excel ../excelgantt/sched.xlsx -sheet Sheet1"
    assertReturnsOKMultiArgs "${EXCELGANTT}" "-v -excel ../excelgantt/color.xlsx -sheet Sheet1"
}

function testSuiteExcelganttReturnsNG() {
    assertReturnsNGMultiArgs "${EXCELGANTT}" "-v -excel nofile.xlsx -sheet Sheet1"
    assertReturnsNGMultiArgs "${EXCELGANTT}" "-v -excel ../excelgantt/sched.xlsx -sheet unknownSheet"
}

function testSuiteExcelgantt() {
    testSuiteExcelganttReturnsOK
    testSuiteExcelganttReturnsNG
}

testSuiteHello
testSuiteCreatemock
testSuiteExcelgo
testSuiteExcelgantt

echo successed
