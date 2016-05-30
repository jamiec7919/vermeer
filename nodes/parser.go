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
	TOK_COMMA
)

var type_Int32 = reflect.TypeOf(int32(0))
var type_UInt32 = reflect.TypeOf(uint32(0))
var type_Vec3 = reflect.TypeOf(m.Vec3{})
var type_Vec2 = reflect.TypeOf(m.Vec2{})
var type_Matrix = reflect.TypeOf(m.Matrix4{})
var type_PointArray = reflect.TypeOf(core.PointArray{})
var type_Vec2Array = reflect.TypeOf(core.Vec2Array{})
var type_Vec3Array = reflect.TypeOf(core.Vec3Array{})
var type_MatrixArray = reflect.TypeOf(core.MatrixArray{})

// The type of symbols returned from lexer (shouldn't be public)
type SymType struct {
	numFloat float64
	numInt   int64
	str      string // might be string const or token

}

type parser struct {
	lex *Lex
	rc  *core.RenderContext
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

	parser := parser{lex: &l, rc: rc}

	//	l.error = parser.error

	return parser.parse()

}

func (p *parser) int32slice(field reflect.Value) error {
	var sym SymType

	count := -1

	if t := p.lex.Lex(&sym); t != TOK_INT {
		return errors.New("Expected slice length.")
	}

	count = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TOK_TOKEN && sym.str != "int" {
		return errors.New("Expected slice type.")
	}

	s := make([]int32, 0, count)

	for i := 0; i < count; i++ {
		if t := p.lex.Lex(&sym); t != TOK_INT {
			return errors.New("Expected int.")
		}

		s = append(s, int32(sym.numInt))
	}

	field.Set(reflect.ValueOf(s))

	return nil
}

func (p *parser) rgb(field reflect.Value) error {

	var sym SymType

	if t := p.lex.Lex(&sym); t != TOK_TOKEN && sym.str != "rgb" {
		return errors.New("Expected field type.")
	}

	v := &core.ConstantMap{}

	for i := range v.C {
		switch t := p.lex.Lex(&sym); t {
		case TOK_INT:
			v.C[i] = float32(sym.numInt)
		case TOK_FLOAT:
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

	if t := p.lex.Lex(&sym); t != TOK_TOKEN && sym.str != "float" {
		return errors.New("Expected field type.")
	}

	v := &core.ConstantMap{}

	switch t := p.lex.Lex(&sym); t {
	case TOK_INT:
		v.C[0] = float32(sym.numInt)
		v.C[1] = float32(sym.numInt)
		v.C[2] = float32(sym.numInt)
	case TOK_FLOAT:
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

	if t := p.lex.Lex(&sym); t != TOK_TOKEN && sym.str != "rgbtex" {
		return errors.New("Expected field type.")
	}

	v := &core.TextureMap{}

	if t := p.lex.Lex(&sym); t != TOK_STRING {
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
		case TOK_INT:
			v[i] = float32(sym.numInt)
		case TOK_FLOAT:
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

	if t := p.lex.Lex(&sym); t != TOK_INT {
		return errors.New("Expected number of motion keys.")
	}

	v.MotionKeys = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TOK_INT {
		return errors.New("Expected number of elements.")
	}

	v.ElemsPerKey = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TOK_TOKEN && sym.str != "point" {
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
			case TOK_INT:
				el[i] = float32(sym.numInt)
			case TOK_FLOAT:
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

	if t := p.lex.Lex(&sym); t != TOK_INT {
		return errors.New("Expected number of motion keys.")
	}

	v.MotionKeys = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TOK_INT {
		return errors.New("Expected number of elements.")
	}

	v.ElemsPerKey = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TOK_TOKEN && sym.str != "vec3" {
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
			case TOK_INT:
				el[i] = float32(sym.numInt)
			case TOK_FLOAT:
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

	if t := p.lex.Lex(&sym); t != TOK_INT {
		return errors.New("Expected number of motion keys.")
	}

	v.MotionKeys = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TOK_INT {
		return errors.New("Expected number of elements.")
	}

	v.ElemsPerKey = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TOK_TOKEN && sym.str != "vec2" {
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
			case TOK_INT:
				el[i] = float32(sym.numInt)
			case TOK_FLOAT:
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

	if t := p.lex.Lex(&sym); t != TOK_INT {
		return errors.New("Expected number of motion keys.")
	}

	v.MotionKeys = int(sym.numInt)

	if t := p.lex.Lex(&sym); t != TOK_TOKEN && sym.str != "matrix" {
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
			case TOK_INT:
				mat[i] = float32(sym.numInt)
			case TOK_FLOAT:
				mat[i] = float32(sym.numFloat)
			default:
				return errors.New("Expected matrix component.")
			}

			v.Elems = append(v.Elems, mat)
		}
	}

	field.Set(reflect.ValueOf(v))

	return nil
}

func (p *parser) matrix(field reflect.Value) error {

	var sym SymType

	if t := p.lex.Lex(&sym); t != TOK_TOKEN && sym.str != "matrix" {
		return errors.New("Expected matrix type.")
	}

	v := m.Matrix4{}

	for i := range v {
		switch t := p.lex.Lex(&sym); t {
		case TOK_INT:
			v[i] = float32(sym.numInt)
		case TOK_FLOAT:
			v[i] = float32(sym.numFloat)
		default:
			return errors.New("Expected matrix component.")
		}

	}

	field.Set(reflect.ValueOf(v))

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
		case TOK_FLOAT:
			field.SetInt(int64(v.numFloat))
		case TOK_INT:
			field.SetInt(v.numInt)
		}

	case reflect.Float32, reflect.Float64:
		switch t := p.lex.Lex(&v); t {
		case TOK_FLOAT:
			field.SetFloat(v.numFloat)
		case TOK_INT:
			field.SetFloat(float64(v.numInt))
		}

	case reflect.Bool:
		switch t := p.lex.Lex(&v); t {
		case TOK_INT:
			if v.numInt == 0 {
				field.SetBool(false)
			} else {
				field.SetBool(true)
			}
		}

	case reflect.String:
		switch t := p.lex.Lex(&v); t {
		case TOK_STRING:
			field.SetString(v.str)
		}

	case reflect.Slice:
		switch field.Type().Elem() {
		case type_Int32:
			switch t := p.lex.Peek(&v); t {
			case TOK_INT:
				p.int32slice(field)
			default:
				p.error("Invalid token for param (expecting length of slice)")
				p.lex.Skip()
			}
		}

	case reflect.Interface:
		if t := p.lex.Peek(&v); t == TOK_TOKEN {
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
		case type_Vec3:
			return p.vec3(field)
		case type_MatrixArray:
			return p.matrixarray(field)
		case type_PointArray:
			return p.pointarray(field)
		case type_Vec3Array:
			return p.vec3array(field)
		case type_Vec2Array:
			return p.vec2array(field)
		default:
			p.error("Invalid type for param (%v)", field.Type())
			p.lex.Skip()
		}

	}

	return nil
}

func lookupParam(node core.Node, field_name string) (reflect.Value, error) {
	rv := reflect.ValueOf(node)

	relem := rv.Elem()

	if relem.Kind() != reflect.Struct {
		return reflect.Value{}, errors.New("node is no a struct")

	}

	ty := relem.Type()

	for i := 0; i < relem.NumField(); i++ {
		f := ty.Field(i)
		if tag := f.Tag.Get("node"); tag != "" {
			if tag == field_name {
				return relem.Field(i), nil

			}
		} else {
			if f.Name == field_name {
				return relem.Field(i), nil

			}

		}

	}

	return reflect.Value{}, errors.New("Field " + field_name + " not found.")
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
		case TOK_TOKEN:
			param_name := v.str

			field, err := lookupParam(node, param_name)

			if err != nil {
				return nil, err
			}

			if !field.IsValid() {
				p.error("Field %v not found/invalid in %v", param_name, name)
				return nil, nil
			}

			if err := p.param(field); err != nil {
				p.error("Error parsing field: %v", err)
			}

		case TOK_CLOSECURLYBRACE:
			//log.Printf("Got obj %v %v", objtype, params)
			return node, nil

		default:
			p.error("parseNode: Error, invalid token in object \"%v\" %v", t, v)

		}
	}

}

func (p *parser) error(msg string, v ...interface{}) {
	if err := p.rc.Error(errors.New(fmt.Sprintf(msg, v...))); err != nil {
		panic(err)
	}
}

func (p *parser) parse() error {
	var v SymType
L:
	for {
		t := p.lex.Lex(&v)
		switch t {
		case TOK_TOKEN:
			token := v.str
			if t := p.lex.Lex(&v); t != TOK_OPENCURLYBRACE {
				p.error("Invalid token in node preamble")
			}

			node, err := p.node(token)

			if err != nil || node == nil {
				p.error("Node is nil: %v", err)

				for {
					t := p.lex.Lex(&v)
					// skip until closing brace
					if t == TOK_CLOSECURLYBRACE {
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
