// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	//"fmt"
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/texture"
)

type Texture struct {
	Filename string
	Chan     int
}

func (c *Texture) Float32(sg *core.ShaderContext) float32 {
	return texture.SampleRGB(c.Filename, sg.U, sg.V, 1, 1)[c.Chan]
}

func (c *Texture) RGB(sg *core.ShaderContext) colour.RGB {
	deltaTx := m.Vec2Scale(sg.Image.PixelDelta[0], sg.Dduvdx)
	deltaTy := m.Vec2Scale(sg.Image.PixelDelta[1], sg.Dduvdy)

	ds := m.Max(m.Abs(deltaTx[0]), m.Abs(deltaTy[0]))
	dt := m.Max(m.Abs(deltaTx[1]), m.Abs(deltaTy[1]))

	ds = m.Vec2Length(deltaTx)
	dt = m.Vec2Length(deltaTy)

	//fmt.Printf("x: %v y: %v\n", deltaTx, deltaTy)
	return colour.RGB(texture.SampleRGB(c.Filename, sg.U, sg.V, ds, dt))

}
