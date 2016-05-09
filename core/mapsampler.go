// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/jamiec7919/vermeer/material/texture"
)

type MapSampler interface {
	SampleRGB(s, t, ds, dt float32) (out [3]float32)
	SampleScalar(s, t, ds, dt float32) (out float32)
}

type ScalarSampler interface {
	Sample(u, v float32) float32
}

type TripleSampler interface {
	Sample(u, v float32) (r, g, b float32)
}

type ConstantMap struct {
	C [3]float32
}

func (c *ConstantMap) SampleRGB(s, t, ds, dt float32) (out [3]float32) {
	out[0] = c.C[0]
	out[1] = c.C[1]
	out[2] = c.C[2]
	return
}

func (c *ConstantMap) SampleScalar(s, t, ds, dt float32) (out float32) {
	out = c.C[0]
	return
}

type TextureMap struct {
	Filename string
}

func (c *TextureMap) SampleRGB(s, t, ds, dt float32) (out [3]float32) {
	return texture.SampleRGB(c.Filename, s, t, ds, dt)
}

func (c *TextureMap) SampleScalar(s, t, ds, dt float32) (out float32) {
	q := texture.SampleRGB(c.Filename, s, t, ds, dt)
	return q[0]

}
