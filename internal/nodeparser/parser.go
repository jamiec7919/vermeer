// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nodeparser

/*
	This is a bit of a mess, roll it up into a struct maybe and clean up error handling.
*/

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	// "os"
	"errors"
	"strconv"
	"unicode"
	"unicode/utf8"
)

const (
	eof = iota
	TOK_TOKEN
	TOK_STRING
	TOK_FLOAT
	TOK_INT
	TOK_OPENBRACE
	TOK_CLOSEBRACE
	TOK_OPENCURLYBRACE
	TOK_CLOSECURLYBRACE
)

// Lex.readLine: Warning: if there is no line end then this returns wrongly that there is an error.

type Params map[string]interface{}

type Dispatcher interface {
	Dispatch(method string, v map[string]interface{}) error

	// Any errors reported by dispatch are passed here, if this returns an error
	// then it is treated as fatal and parser stops.
	Error(error) error
}

// The parser expects the lexer to return 0 on EOF.  Give it a name
// for clarity.
//const eof = 0

type Lex struct {
	line []byte
	peek rune
	in   *bufio.Reader
}

type SymType struct {
	numFloat float64
	numInt   int64
	str      string // might be string const or token

}

func _error(disp Dispatcher, msg string) {
	if err := disp.Error(errors.New(msg)); err != nil {
		panic(err)
	}
}

func isAlpha(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// The parser calls this method to get each new token.  This
// implementation returns operators and NUM.
func (x *Lex) Lex(yylval *SymType) int {
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

		case ' ', '\t', '\r', '\n', '"', '[', ']', '{', '}':
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

const (
	array_NONE = iota
	array_INTARRAY
	array_FLOATARRAY
	array_STRINGARRAY
	array_OBJARRAY
)

func parseArray(l *Lex) (interface{}, error) {

	var v SymType

	arrayType := array_NONE
	var fArray []float64
	var iArray []int64
	var sArray []string
	var pArray []Params

	for {
		t := l.Lex(&v)
		switch t {
		case TOK_STRING:
			switch arrayType {
			case array_NONE:
				arrayType = array_STRINGARRAY
				fallthrough
			case array_STRINGARRAY:
				sArray = append(sArray, v.str)
			default:
				return nil, nil // should return error
			}
			//   log.Printf("STRING: %v",v.str)

		case TOK_FLOAT:
			switch arrayType {
			case array_INTARRAY:
				// convert to float array
				for i := range iArray {
					fArray = append(fArray, float64(iArray[i]))
				}
				iArray = nil
				fallthrough
			case array_NONE:
				arrayType = array_FLOATARRAY
				fallthrough
			case array_FLOATARRAY:
				fArray = append(fArray, v.numFloat)
			default:
				return nil, nil // should return error
			}
			// log.Printf("FLOAT: %v",v.numFloat)
		case TOK_INT:
			switch arrayType {
			case array_NONE:
				arrayType = array_INTARRAY
				fallthrough
			case array_INTARRAY:
				iArray = append(iArray, v.numInt)
			case array_FLOATARRAY:
				fArray = append(fArray, float64(v.numInt))
			default:
				return nil, nil // should return error
			}
			//log.Printf("INT: %v",v.numInt)

		case TOK_OPENCURLYBRACE:
			// log.Printf("BEGIN OBJ (array)")
			params, _ := parseObj(l)

			pArray = append(pArray, params)

			switch arrayType {
			case array_NONE:
				arrayType = array_OBJARRAY
			default:
				// log.Printf("ARray err")
				return nil, nil // should return error
			}

		case TOK_CLOSEBRACE:
			// .Printf("]")
			switch arrayType {
			case array_FLOATARRAY:
				return fArray, nil
			case array_INTARRAY:
				return iArray, nil
			case array_STRINGARRAY:
				return sArray, nil
			case array_OBJARRAY:
				return pArray, nil
			default:
				return nil, nil
			}
		default:
			return nil, nil // ERROR
		}
	}

}

func parseParam(l *Lex) (param interface{}, err error) {

	var v SymType

	for {
		t := l.Lex(&v)
		switch t {
		case TOK_FLOAT:
			return v.numFloat, nil
			//log.Printf("FLOAT: %v",v.numFloat)
		case TOK_INT:
			return v.numInt, nil
			//log.Printf("INT: %v",v.numInt)
		case TOK_OPENBRACE:
			// collect items until closebrace.
			// Arrays either arrays of strings or arrays of
			// numbers.  Integers are promoted to floats if any element of
			// array is a float.
			return parseArray(l)
		case TOK_STRING:
			return v.str, nil

		}
	}

	return nil, nil
}

func parseObj(l *Lex) (params map[string]interface{}, err error) {
	params = map[string]interface{}{}

	var v SymType

	for {
		t := l.Lex(&v)
		// log.Printf("%v", t)
		switch t {
		case TOK_TOKEN:
			token := v.str

			param, _ := parseParam(l)

			params[token] = param

		case TOK_CLOSECURLYBRACE:
			// log.Printf("Got obj %v", params)
			return

		default:
			log.Printf("parseObj: Error, invalid token in object \"%v\" %v", t, v)

		}
	}

}

func Parse(disp Dispatcher, filename string) error {

	b, err := ioutil.ReadFile(filename)
	/*
		f, err := os.Open(filename)
	*/
	if err != nil {
		return err

	}
	/*
		in := bufio.NewReader(f)
	*/
	in := bufio.NewReader(bytes.NewBuffer(b))
	var v SymType
	var l Lex
	l.in = in

	var token string
L:
	for {
		t := l.Lex(&v)
		switch t {
		/* case TOK_STRING:
		//log.Printf("STRING: %v",v.str)
		args = append(args, v.str)*/
		case TOK_TOKEN:
			token = v.str

			if t := l.Lex(&v); t != TOK_OPENCURLYBRACE {
				_error(disp, "Invalid token in obj preamble")
			}

			params, _ := parseObj(&l)
			/*
				if token != "" {
					// process token and args
					doToken(token, args, cxt)
				}
				token = v.str
				args = nil */

			//log.Printf("Got Object %v %v", token, params)
			if err := disp.Dispatch(token, params); err != nil {
				if err2 := disp.Error(err); err2 != nil {
					return err2
				}
			}
			// if current token is not nil then output
			//log.Printf("TOKEN: %v",v.str)
		/* case TOK_FLOAT:
			args = append(args, v.numFloat)
			//log.Printf("FLOAT: %v",v.numFloat)
		case TOK_INT:
			args = append(args, v.numInt)
			//log.Printf("INT: %v",v.numInt)
		case TOK_OPENBRACE:
			// collect items until closebrace.
			// Arrays either arrays of strings or arrays of
			// numbers.  Integers are promoted to floats if any element of
			// array is a float.
			a, _ := parseArray(&l)
		*/
		/* switch a.(type) {
			case []float64:
			  log.Printf("[]float64")
		} */

		//	args = append(args, a)

		//log.Printf("[")
		// case TOK_CLOSECURLYBRACE:
		//log.Printf("]")
		// ERROR
		default:
			break L
		}
	}

	// if token != "" {
	// process token and args
	// 	doToken(token, args, cxt)
	// }
	//}
	return nil
}

/*
func main() {
	f,err := os.Open("test.vi")

	if err != nil {
	  log.Fatalf("Failed to open file: %v",err)

	}

	in := bufio.NewReader(f)

        var v SymType
        var l Lex

	for {
		/*if _, err := os.Stdout.WriteString("> "); err != nil {
			log.Fatalf("WriteString: %s", err)
		}*/ /*
		line, err := in.ReadBytes('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("ReadBytes: %s", err)
		}
                log.Printf(": %v",line)

                l.line = line

                L: for {
                  t := l.Lex(&v)
                  switch t {
                    case TOK_STRING:
                      log.Printf("STRING: %v",v.str)
                    case TOK_TOKEN:
                      log.Printf("TOKEN: %v",v.str)
                    case TOK_FLOAT:
                      log.Printf("FLOAT: %v",v.numFloat)
                    case TOK_INT:
                      log.Printf("INT: %v",v.numInt)
                    case TOK_OPENBRACE:
                      log.Printf("[")
                    case TOK_CLOSEBRACE:
                      log.Printf("]")
                    default:
                      break L
                  }
                }

	}
}
*/
