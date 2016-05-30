// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nodes

/*
	This is a bit of a mess, roll it up into a struct maybe and clean up error handling.
*/

import (
	"bufio"
	"bytes"
	"io"
	"log"
	// "os"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// Lex represents a lexical analyser for the vnf parser. (shouldn't be public)
type Lex struct {
	line []byte
	peek rune
	in   *bufio.Reader

	peekToken bool // Should we pass the token back?
	psym      SymType
	ptoken    int
}

func isAlpha(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// Call this to skip the next token
func (x *Lex) Skip() {
	var sym SymType
	x.Lex(&sym)
}

// Call this to peek ahead at next symbol, the next call to Lex will return same sym.
func (x *Lex) Peek(yylval *SymType) int {
	token := x.lex(yylval)
	x.peekToken = true
	x.psym = *yylval
	x.ptoken = token

	return token
}

// The parser calls this method to get each new token.
func (x *Lex) Lex(yylval *SymType) int {
	if x.peekToken {
		x.peekToken = false
		*yylval = x.psym
		return x.ptoken
	}

	return x.lex(yylval)
}

func (x *Lex) lex(yylval *SymType) int {

	for {
		c := x.next()

		if isAlpha(c) {
			return x.token(c, yylval)
		}

		switch c {
		case eof:
			return eof
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return x.num(c, yylval)
		case '[':
			return TOK_OPENBRACE
		case ']':
			return TOK_CLOSEBRACE
		case '{':
			return TOK_OPENCURLYBRACE
		case '}':
			return TOK_CLOSECURLYBRACE
		case ',':
			return TOK_COMMA
		case '"':
			return x.str(c, yylval)
		case ' ', '\t', '\n': /* Whitespace */

		}

	}
}

func (x *Lex) str(c rune, yylval *SymType) int {
	add := func(b *bytes.Buffer, c rune) {
		if _, err := b.WriteRune(c); err != nil {
			log.Fatalf("WriteRune: %s", err)
		}
	}
	var b bytes.Buffer

	// add(&b, c)
L:
	for {
		c = x.next()
		switch c {
		case eof:
			return eof // error really

		case '"':
			break L
		default:
			add(&b, c)
		}
	}

	yylval.str = b.String()
	return TOK_STRING
}

func (x *Lex) token(c rune, yylval *SymType) int {
	add := func(b *bytes.Buffer, c rune) {
		if _, err := b.WriteRune(c); err != nil {
			log.Fatalf("WriteRune: %s", err)
		}
	}
	var b bytes.Buffer

	add(&b, c)
L:
	for {
		c = x.next()
		switch c {
		case eof:
			return eof // error really
		case '"', '[', ']', '{', '}', ',':
			x.peek = c
			break L
		case ' ', '\t', '\r', '\n':
			break L
		default:
			add(&b, c)
		}
	}

	yylval.str = b.String()
	return TOK_TOKEN
}

func (x *Lex) num(c rune, yylval *SymType) int {
	add := func(b *bytes.Buffer, c rune) {
		if _, err := b.WriteRune(c); err != nil {
			log.Fatalf("WriteRune: %s", err)
		}
	}
	var b bytes.Buffer
	isFloat := false
	add(&b, c)
L:
	for {
		c = x.next()
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
			add(&b, c)
		case '.', 'e', 'E':
			isFloat = true
			add(&b, c)
		default:
			break L
		}
	}
	if c != eof {
		x.peek = c
	}

	if isFloat {
		f, err := strconv.ParseFloat(b.String(), 64)

		if err != nil {
			return eof
		}
		yylval.numFloat = f
		return TOK_FLOAT
	} else {
		f, err := strconv.ParseInt(b.String(), 10, 64)

		if err != nil {
			return eof
		}
		yylval.numInt = f
		return TOK_INT
	}

}

func (x *Lex) readLine() error {
	line, err := x.in.ReadBytes('\n')
	if err == io.EOF {
		if len(line) < 1 { // last line may not end with \n so detect with empty data
			return err
		}
	}
	x.line = line
	return nil
}

// Return the next rune for the lexer.
func (x *Lex) next() rune {
	if x.peek != eof {
		r := x.peek
		x.peek = eof
		return r
	}
	if len(x.line) == 0 {
		err := x.readLine()

		if err == io.EOF {
			return eof
		}

		return x.next()
	}
	c, size := utf8.DecodeRune(x.line)
	x.line = x.line[size:]
	if c == utf8.RuneError && size == 1 {
		log.Print("invalid utf8")
		return x.next()
	}

	if c == '#' {
		err := x.readLine()

		if err == io.EOF {
			return eof
		}

		return '\n'

	}

	return c
}
