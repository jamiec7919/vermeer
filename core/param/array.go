// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package param

import (
	m "github.com/jamiec7919/vermeer/math"
)

/*
 These wrap the idea of arrays of elements (e.g. points, vec2's or matrices) for motion keys.
 storage is simply an array of the appropriate type, of length MotionKeys*ElemsPerKey (for types
 where that makes sense).
*/

// PointArray represents an set of motion keys, with ElemsPerKey points per key.
// Points are represented by Vec3.
type PointArray struct {
	MotionKeys  int // Number of motion keys
	ElemsPerKey int // Number of elements per key
	Elems       []m.Vec3
}

// Key returns the slice of elements for the given key
func (a *PointArray) Key(key int) []m.Vec3 {
	return a.Elems[key*a.ElemsPerKey : (key+1)*a.ElemsPerKey]
}

// Vec3Array represents an set of motion keys, with ElemsPerKey vec3s per key.
type Vec3Array struct {
	MotionKeys  int // Number of motion keys
	ElemsPerKey int // Number of elements per key
	Elems       []m.Vec3
}

// Key returns the slice of elements for the given key
func (a *Vec3Array) Key(key int) []m.Vec3 {
	return a.Elems[key*a.ElemsPerKey : (key+1)*a.ElemsPerKey]
}

// Vec2Array represents an set of motion keys, with ElemsPerKey vec2 per key.
type Vec2Array struct {
	MotionKeys  int // Number of motion keys
	ElemsPerKey int // Number of elements per key
	Elems       []m.Vec2
}

// Key returns the slice of elements for the given key
func (a *Vec2Array) Key(key int) []m.Vec2 {
	return a.Elems[key*a.ElemsPerKey : (key+1)*a.ElemsPerKey]
}

// MatrixArray represents an set of motion keys, one Matrix4 per key.
type MatrixArray struct {
	MotionKeys int // Number of motion keys
	Elems      []m.Matrix4
}

// Float32Array represents an set of motion keys, with ElemsPerKey points per key.
type Float32Array struct {
	MotionKeys  int // Number of motion keys
	ElemsPerKey int // Number of elements per key
	Elems       []float32
}

// Key returns the slice of elements for the given key
func (a *Float32Array) Key(key int) []float32 {
	return a.Elems[key*a.ElemsPerKey : (key+1)*a.ElemsPerKey]
}
