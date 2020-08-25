#!/bin/sh

BIN_DIR=../bin
HELLO=${BIN_DIR}/hello
YACC=${BIN_DIR}/yacc

${HELLO}
${YACC} "1 + 2"
