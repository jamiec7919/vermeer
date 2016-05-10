// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/material/texture"
)

type Float32Param interface {
	Float32(sg *ShaderGlobals) float32
}

type RGBParam interface {
	RGB(sg *ShaderGlobals) colour.RGB
}

type ConstantMap struct {
	C [3]float32
}

func (c *ConstantMap) RGB(sg *ShaderGlobals) (out colour.RGB) {
	out[0] = c.C[0]
	out[1] = c.C[1]
	out[2] = c.C[2]
	return
}

func (c *ConstantMap) Float32(sg *ShaderGlobals) (out float32) {
	out = c.C[0]
	return
}

type TextureMap struct {
	Filename string
}

func (c *TextureMap) RGB(sg *ShaderGlobals) (out colour.RGB) {
	return colour.RGB(texture.SampleRGB(c.Filename, sg.U, sg.V, 1, 1))
}

func (c *TextureMap) Float32(sg *ShaderGlobals) (out float32) {
	q := texture.SampleRGB(c.Filename, sg.U, sg.V, 1, 1)
	return q[0]

}
