// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package nodes is used to parse vnf files and create the node structure.
*/
package nodes

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"os"
	"reflect"
)

// Token types returned by lexer.
const (
	eof = iota
	TokToken
	TokString
	TokFloat
	TokInt
	TokOpenBrace
	TokCloseBrace
	TokOpenCurlyBrace
	TokCloseCurlyBrace
	TokComma
)

var typeInt32 = reflect.TypeOf(int32(0))
var typeUInt32 = reflect.TypeOf(uint32(0))
var typeVec3 = reflect.TypeOf(m.Vec3{})
var typeVec2 = reflect.TypeOf(m.Vec2{})
var typeMatrix = reflect.TypeOf(m.Matrix4{})
var typePointArray = reflect.TypeOf(core.PointArray{})
var typeVec2Array = reflect.TypeOf(core.Vec2Array{})
var typeVec3Array = reflect.TypeOf(core.Vec3Array{})
var typeFloat32Array = reflect.TypeOf(core.Float32Array{})
var typeMatrixArray = reflect.TypeOf(core.MatrixArray{})

// SymType is the type of symbols returned from lexer (shouldn't be public)
type SymType struct {
	numFloat float64
	numInt   int64
	str      string // might be string const or token

}

type parser struct {
	filename string
	lex      *Lex
	rc       *core.RenderContext
}

func init() {
	Register("Globals", func() (core.Node, error) {

		return &core.Globals{XRes: 256, YRes: 256, MaxGoRoutines: core.MAXGOROUTINES}, nil
	})
}

// Parse attempts to open filename and parse the contents, adding nodes to rc.  Returns
// nil on success or an appropriate error.
func Parse(rc *core.RenderContext, filename string) error {

	f, err := os.Open(filename)

	if err != nil {
		return err

	}

	in := bufio.NewReader(f)

	var l Lex
	l.in = in
	l.LineNumber = 0
	l.ColNumber = 1

	parser := parser{filename: filename, lex: &l, rc: rc}

	//	l.error = parser.error

	return parser.parse()

}

func (p *parser) int32slice(field reflect.Value) error {
	var sym SymType

	count := -1

	if t := p.lex.Lex(&sym); t != TokInt {
		return errors.New("Expected slice length.")
	}

	count = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TokToken && sym.str != "int" {
		return errors.New("Expected slice type.")
	}

	s := make([]int32, 0, count)

	for i := 0; i < count; i++ {
		if t := p.lex.Lex(&sym); t != TokInt {
			return errors.New("Expected int.")
		}

		s = append(s, int32(sym.numInt))
	}

	field.Set(reflect.ValueOf(s))

	return nil
}

func (p *parser) rgb(field reflect.Value) error {

	var sym SymType

	if t := p.lex.Lex(&sym); t != TokToken && sym.str != "rgb" {
		return errors.New("Expected field type.")
	}

	v := &core.ConstantMap{}

	for i := range v.C {
		switch t := p.lex.Lex(&sym); t {
		case TokInt:
			v.C[i] = float32(sym.numInt)
		case TokFloat:
			v.C[i] = float32(sym.numFloat)
		default:
			return errors.New("Expected RGB component.")
		}

	}

	field.Set(reflect.ValueOf(v))

	return nil
}

func (p *parser) floatmap(field reflect.Value) error {

	var sym SymType

	if t := p.lex.Lex(&sym); t != TokToken && sym.str != "float" {
		return errors.New("Expected field type.")
	}

	v := &core.ConstantMap{}

	switch t := p.lex.Lex(&sym); t {
	case TokInt:
		v.C[0] = float32(sym.numInt)
		v.C[1] = float32(sym.numInt)
		v.C[2] = float32(sym.numInt)
	case TokFloat:
		v.C[0] = float32(sym.numFloat)
		v.C[1] = float32(sym.numFloat)
		v.C[2] = float32(sym.numFloat)
	default:
		return errors.New("Expected float component.")
	}

	field.Set(reflect.ValueOf(v))

	return nil
}

func (p *parser) rgbtex(field reflect.Value) error {

	var sym SymType

	if t := p.lex.Lex(&sym); t != TokToken && sym.str != "rgbtex" {
		return errors.New("Expected field type.")
	}

	v := &core.TextureMap{}

	if t := p.lex.Lex(&sym); t != TokString {
		return errors.New("Expected RGB texture filename.")
	}

	v.Filename = sym.str

	field.Set(reflect.ValueOf(v))

	return nil
}

func (p *parser) vec3(field reflect.Value) error {

	var sym SymType

	v := m.Vec3{}

	for i := range v {
		switch t := p.lex.Lex(&sym); t {
		case TokInt:
			v[i] = float32(sym.numInt)
		case TokFloat:
			v[i] = float32(sym.numFloat)
		default:
			return errors.New("Expected vector component.")
		}

	}

	field.Set(reflect.ValueOf(v))

	return nil
}

func (p *parser) pointarray(field reflect.Value) error {

	var sym SymType

	v := core.PointArray{}

	if t := p.lex.Lex(&sym); t != TokInt {
		return errors.New("Expected number of motion keys.")
	}

	v.MotionKeys = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TokInt {
		return errors.New("Expected number of elements.")
	}

	v.ElemsPerKey = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TokToken && sym.str != "point" {
		return errors.New("Expected array type.")
	}

	k := v.MotionKeys

	if k == 0 {
		k = 1
	}

	v.Elems = make([]m.Vec3, 0, k*v.ElemsPerKey)

	for j := 0; j < k*v.ElemsPerKey; j++ {
		el := m.Vec3{}

		for i := range el {
			switch t := p.lex.Lex(&sym); t {
			case TokInt:
				el[i] = float32(sym.numInt)
			case TokFloat:
				el[i] = float32(sym.numFloat)
			default:
				return errors.New("Expected point component.")
			}

		}
		v.Elems = append(v.Elems, el)
	}

	field.Set(reflect.ValueOf(v))

	return nil
}

func (p *parser) vec3array(field reflect.Value) error {

	var sym SymType

	v := core.Vec3Array{}

	if t := p.lex.Lex(&sym); t != TokInt {
		return errors.New("Expected number of motion keys.")
	}

	v.MotionKeys = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TokInt {
		return errors.New("Expected number of elements.")
	}

	v.ElemsPerKey = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TokToken && sym.str != "vec3" {
		return errors.New("Expected array type.")
	}

	k := v.MotionKeys

	if k == 0 {
		k = 1
	}

	v.Elems = make([]m.Vec3, 0, k*v.ElemsPerKey)

	for j := 0; j < k*v.ElemsPerKey; j++ {
		el := m.Vec3{}

		for i := range el {
			switch t := p.lex.Lex(&sym); t {
			case TokInt:
				el[i] = float32(sym.numInt)
			case TokFloat:
				el[i] = float32(sym.numFloat)
			default:
				return errors.New("Expected vec3 component.")
			}

		}
		v.Elems = append(v.Elems, el)
	}

	field.Set(reflect.ValueOf(v))

	return nil
}

func (p *parser) vec2array(field reflect.Value) error {

	var sym SymType

	v := core.Vec2Array{}

	if t := p.lex.Lex(&sym); t != TokInt {
		return errors.New("Expected number of motion keys.")
	}

	v.MotionKeys = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TokInt {
		return errors.New("Expected number of elements.")
	}

	v.ElemsPerKey = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TokToken && sym.str != "vec2" {
		return errors.New("Expected array type.")
	}

	k := v.MotionKeys

	if k == 0 {
		k = 1
	}

	v.Elems = make([]m.Vec2, 0, k*v.ElemsPerKey)

	for j := 0; j < k*v.ElemsPerKey; j++ {
		el := m.Vec2{}

		for i := range el {
			switch t := p.lex.Lex(&sym); t {
			case TokInt:
				el[i] = float32(sym.numInt)
			case TokFloat:
				el[i] = float32(sym.numFloat)
			default:
				return errors.New("Expected vec2 component.")
			}

		}
		v.Elems = append(v.Elems, el)
	}

	field.Set(reflect.ValueOf(v))

	return nil
}

func (p *parser) matrixarray(field reflect.Value) error {

	var sym SymType

	v := core.MatrixArray{}

	if t := p.lex.Lex(&sym); t != TokInt {
		return errors.New("Expected number of motion keys.")
	}

	v.MotionKeys = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TokToken && sym.str != "matrix" {
		return errors.New("Expected array type.")
	}

	k := v.MotionKeys

	if k == 0 {
		k = 1
	}

	v.Elems = make([]m.Matrix4, 0, k)

	for j := 0; j < k; j++ {
		mat := m.Matrix4{}

		for i := range mat {
			switch t := p.lex.Lex(&sym); t {
			case TokInt:
				mat[i] = float32(sym.numInt)
			case TokFloat:
				mat[i] = float32(sym.numFloat)
			default:
				return errors.New("Expected matrix component.")
			}

			v.Elems = append(v.Elems, m.Matrix4Transpose(mat))
		}
	}

	field.Set(reflect.ValueOf(v))

	return nil
}

func (p *parser) float32array(field reflect.Value) error {

	var sym SymType

	v := core.Float32Array{}

	if t := p.lex.Lex(&sym); t != TokInt {
		return errors.New("Expected number of motion keys.")
	}

	v.MotionKeys = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TokInt {
		return errors.New("Expected number of elements.")
	}

	v.ElemsPerKey = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TokToken && sym.str != "float" {
		return errors.New("Expected array type.")
	}

	k := v.MotionKeys

	if k == 0 {
		k = 1
	}

	v.Elems = make([]float32, 0, k*v.ElemsPerKey)

	for j := 0; j < k*v.ElemsPerKey; j++ {
		var el float32

		switch t := p.lex.Lex(&sym); t {
		case TokInt:
			el = float32(sym.numInt)
		case TokFloat:
			el = float32(sym.numFloat)
		default:
			return errors.New("Expected float32 element.")
		}

		v.Elems = append(v.Elems, el)
	}

	field.Set(reflect.ValueOf(v))

	return nil
}

func (p *parser) matrix(field reflect.Value) error {

	var sym SymType

	if t := p.lex.Lex(&sym); t != TokToken && sym.str != "matrix" {
		return errors.New("Expected matrix type.")
	}

	v := m.Matrix4{}

	for i := range v {
		switch t := p.lex.Lex(&sym); t {
		case TokInt:
			v[i] = float32(sym.numInt)
		case TokFloat:
			v[i] = float32(sym.numFloat)
		default:
			return errors.New("Expected matrix component.")
		}

	}

	field.Set(reflect.ValueOf(m.Matrix4Transpose(v)))

	return nil
}

func (p *parser) param(field reflect.Value) error {

	if !field.CanSet() {
		return errors.New("Can't set field")
	}

	var v SymType

	switch field.Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		switch t := p.lex.Lex(&v); t {
		case TokFloat:
			field.SetInt(int64(v.numFloat))
		case TokInt:
			field.SetInt(v.numInt)
		}

	case reflect.Float32, reflect.Float64:
		switch t := p.lex.Lex(&v); t {
		case TokFloat:
			field.SetFloat(v.numFloat)
		case TokInt:
			field.SetFloat(float64(v.numInt))
		}

	case reflect.Bool:
		switch t := p.lex.Lex(&v); t {
		case TokInt:
			if v.numInt == 0 {
				field.SetBool(false)
			} else {
				field.SetBool(true)
			}
		}

	case reflect.String:
		switch t := p.lex.Lex(&v); t {
		case TokString:
			field.SetString(v.str)
		}

	case reflect.Slice:
		switch field.Type().Elem() {
		case typeInt32:
			switch t := p.lex.Peek(&v); t {
			case TokInt:
				p.int32slice(field)
			default:
				p.errorf("Invalid token for param (expecting length of slice)")
				p.lex.Skip()
			}
		}

	case reflect.Interface:
		if t := p.lex.Peek(&v); t == TokToken {
			switch v.str {
			case "rgb":
				p.rgb(field)
			case "float":
				p.floatmap(field)
			case "rgbtex":
				p.rgbtex(field)
			}
		}
	default:

		switch field.Type() {
		case typeVec3:
			return p.vec3(field)
		case typeMatrixArray:
			return p.matrixarray(field)
		case typePointArray:
			return p.pointarray(field)
		case typeVec3Array:
			return p.vec3array(field)
		case typeVec2Array:
			return p.vec2array(field)
		case typeFloat32Array:
			return p.float32array(field)
		default:
			p.errorf("Invalid type for param (%v)", field.Type())
			p.lex.Skip()
		}

	}

	return nil
}

func lookupParam(node core.Node, fieldName string) (reflect.Value, error) {
	rv := reflect.ValueOf(node)

	relem := rv.Elem()

	if relem.Kind() != reflect.Struct {
		return reflect.Value{}, errors.New("node is no a struct")

	}

	ty := relem.Type()

	for i := 0; i < relem.NumField(); i++ {
		f := ty.Field(i)
		if tag := f.Tag.Get("node"); tag != "" {
			if tag == fieldName {
				return relem.Field(i), nil

			}
		} else {
			if f.Name == fieldName {
				return relem.Field(i), nil

			}

		}

	}

	return reflect.Value{}, errors.New("Field " + fieldName + " not found.")
}

func (p *parser) node(name string) (core.Node, error) {

	var v SymType

	node, err := createNode(name)

	if err != nil || node == nil {

		return nil, err
	}

	for {
		t := p.lex.Lex(&v)
		// log.Printf("%v", t)
		switch t {
		case TokToken:
			paramName := v.str

			field, err := lookupParam(node, paramName)

			if err != nil {
				return nil, err
			}

			if !field.IsValid() {
				p.errorf("Field %v not found/invalid in %v", paramName, name)
				return nil, nil
			}

			if err := p.param(field); err != nil {
				p.errorf("Error parsing field: %v", err)
			}

		case TokCloseCurlyBrace:
			//log.Printf("Got obj %v %v", objtype, params)
			return node, nil

		default:
			p.errorf("parseNode: Error, invalid token in object \"%v\" %v", t, v)

		}
	}

}

func (p *parser) errorf(msg string, v ...interface{}) {
	line := p.lex.LineNumber
	col := p.lex.ColNumber
	if err := p.rc.Error(fmt.Errorf("%v:%v:%v: %v", p.filename, line, col, fmt.Sprintf(msg, v...))); err != nil {
		panic(err)
	}
}

func (p *parser) parse() error {
	var v SymType
L:
	for {
		t := p.lex.Lex(&v)
		switch t {
		case TokToken:
			token := v.str
			if t := p.lex.Lex(&v); t != TokOpenCurlyBrace {
				p.errorf("Invalid token in node preamble")
			}

			node, err := p.node(token)

			if err != nil || node == nil {
				p.errorf("Node is nil: %v", err)

				for {
					t := p.lex.Lex(&v)
					// skip until closing brace
					if t == TokCloseCurlyBrace {
						break
					}
				}

			}

			if node != nil {
				p.rc.AddNode(node)
			}

		// ERROR
		default:
			break L
		}
	}

	return nil
}
