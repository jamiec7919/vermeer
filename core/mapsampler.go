// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/material/texture"
)

// Float32Param is used by shaders for parseable parameters. This represents a float32 param
// which may be constant, from a texture or something more complex.
type Float32Param interface {
	Float32(sg *ShaderGlobals) float32
}

// RGBParam is used by shaders for parseable parameters. This represents an RGB colour param
// which may be constant, from a texture or something more complex.
type RGBParam interface {
	RGB(sg *ShaderGlobals) colour.RGB
}

// ConstantMap is a concrete parameter type used for shader parameters. This represents a constant
// RGB colour.  Float32 parameters return the red component only.
type ConstantMap struct {
	C [3]float32
}

// RGB implements method for RGBParam.
func (c *ConstantMap) RGB(sg *ShaderGlobals) (out colour.RGB) {
	out[0] = c.C[0]
	out[1] = c.C[1]
	out[2] = c.C[2]
	return
}

// Float32 implements method for Float32Param.
func (c *ConstantMap) Float32(sg *ShaderGlobals) (out float32) {
	out = c.C[0]
	return
}

// TextureMap is a concrete parameter type used for shader parameters. This represents an image file
// RGB 2D texture.  Float32 parameters return the red component only.
type TextureMap struct {
	Filename string
}

// RGB implements method for RGBParam.
func (c *TextureMap) RGB(sg *ShaderGlobals) (out colour.RGB) {
	return colour.RGB(texture.SampleRGB(c.Filename, sg.U, sg.V, 1, 1))
}

// Float32 implements method for Float32Param.
func (c *TextureMap) Float32(sg *ShaderGlobals) (out float32) {
	q := texture.SampleRGB(c.Filename, sg.U, sg.V, 1, 1)
	return q[0]

}
