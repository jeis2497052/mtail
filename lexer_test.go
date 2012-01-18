// Copyright 2011 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package main

import (
	"reflect"
	"strings"
	"testing"
)

type lexerTest struct {
	name   string
	input  string
	tokens []Token
}

var lexerTests = []lexerTest{
	{"empty", "", []Token{
		Token{EOF, "", Position{0, 0, 0}}}},
	{"spaces", " \t\n", []Token{
		Token{EOF, "", Position{1, 0, 0}}}},
	{"comment", "# comment", []Token{
		Token{EOF, "", Position{0, 9, 9}}}},
	{"comment not at col 1", "  # comment", []Token{
		Token{EOF, "", Position{0, 11, 11}}}},
	{"punctuation", "{}(),", []Token{
		Token{LCURLY, "{", Position{0, 0, 0}},
		Token{RCURLY, "}", Position{0, 1, 1}},
		Token{LPAREN, "(", Position{0, 2, 2}},
		Token{RPAREN, ")", Position{0, 3, 3}},
		Token{COMMA, ",", Position{0, 4, 4}},
		Token{EOF, "", Position{0, 5, 5}}}},
	{"keywords",
		"inc\ntag\nstrptime\n", []Token{
			Token{BUILTIN, "inc", Position{0, 0, 2}},
			Token{BUILTIN, "tag", Position{1, 0, 2}},
			Token{BUILTIN, "strptime", Position{2, 0, 7}},
			Token{EOF, "", Position{3, 0, 0}}}},
	{"identifer", "a be foo\nquux line-count", []Token{
		Token{ID, "a", Position{0, 0, 0}},
		Token{ID, "be", Position{0, 2, 3}},
		Token{ID, "foo", Position{0, 5, 7}},
		Token{ID, "quux", Position{1, 0, 3}},
		Token{ID, "line-count", Position{1, 5, 14}},
		Token{EOF, "", Position{1, 15, 15}}}},
	{"regex", "/asdf/", []Token{
		Token{REGEX, "asdf", Position{0, 0, 5}},
		Token{EOF, "", Position{0, 6, 6}}}},
	{"capref", "$foo", []Token{
		Token{CAPREF, "foo", Position{0, 0, 3}},
		Token{EOF, "", Position{0, 4, 4}}}},
	{"numerical capref", "$1", []Token{
		Token{CAPREF, "1", Position{0, 0, 1}},
		Token{EOF, "", Position{0, 2, 2}}}},
	{"capref with trailing punc", "$foo,", []Token{
		Token{CAPREF, "foo", Position{0, 0, 3}},
		Token{COMMA, ",", Position{0, 4, 4}},
		Token{EOF, "", Position{0, 5, 5}}}},
	{"quoted string", "\"asdf\"", []Token{
		Token{STRING, "asdf", Position{0, 0, 5}},
		Token{EOF, "", Position{0, 6, 6}}}},
	{"escaped slashes in regex", "/foo\\//", []Token{
		Token{REGEX, "foo\\/", Position{0, 0, 6}},
		Token{EOF, "", Position{0, 7, 7}}}},
	{"escaped quote in quoted string", "\"\\\"\"", []Token{
		Token{STRING, "\\\"", Position{0, 0, 3}},
		Token{EOF, "", Position{0, 4, 4}}}},
	{"large program",
		"/(?P<date>[[:digit:]-\\/ ])/ {\n" +
			"  strptime($date, \"%Y/%m/%d %H:%M:%S\")\n" +
			"  inc(foo)\n" +
			"}", []Token{
			Token{REGEX, "(?P<date>[[:digit:]-\\/ ])", Position{0, 0, 26}},
			Token{LCURLY, "{", Position{0, 28, 28}},
			Token{BUILTIN, "strptime", Position{1, 2, 9}},
			Token{LPAREN, "(", Position{1, 10, 10}},
			Token{CAPREF, "date", Position{1, 11, 15}},
			Token{COMMA, ",", Position{1, 16, 16}},
			Token{STRING, "%Y/%m/%d %H:%M:%S", Position{1, 18, 36}},
			Token{RPAREN, ")", Position{1, 37, 37}},
			Token{BUILTIN, "inc", Position{2, 2, 4}},
			Token{LPAREN, "(", Position{2, 5, 5}},
			Token{ID, "foo", Position{2, 6, 8}},
			Token{RPAREN, ")", Position{2, 9, 9}},
			Token{RCURLY, "}", Position{3, 0, 0}},
			Token{EOF, "", Position{3, 1, 1}}}},
	// errors
	{"unexpected char", "?", []Token{
		Token{INVALID, "Unexpected input: '?'", Position{0, 0, 0}}}},
	{"unterminated regex", "/foo\n", []Token{
		Token{INVALID, "Unterminated regular expression: \"/foo\"", Position{0, 0, 3}}}},
	{"unterminated quoted string", "\"foo\n", []Token{
		Token{INVALID, "Unterminated quoted string: \"\\\"foo\"", Position{0, 0, 3}}}},
}

// collect gathers the emitted items into a slice.
func collect(t *lexerTest) (tokens []Token) {
	l := NewLexer(t.name, strings.NewReader(t.input))
	for {
		token := l.NextToken()
		tokens = append(tokens, token)
		if token.kind == EOF || token.kind == INVALID {
			break
		}
	}
	return
}

func TestLex(t *testing.T) {
	for _, test := range lexerTests {
		tokens := collect(&test)
		if !reflect.DeepEqual(tokens, test.tokens) {
			t.Errorf("%s: got\n\t%v\nexpected\n\t%v", test.name, tokens, test.tokens)
		}
	}
}