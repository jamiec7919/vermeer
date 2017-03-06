// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	//"fmt"
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/texture"
)

type Texture struct {
	Filename string
	Chan     int
}

func (c *Texture) Float32(sg *core.ShaderContext) float32 {
	return texture.SampleRGB(c.Filename, sg)[c.Chan]
}

func (c *Texture) RGB(sg *core.ShaderContext) colour.RGB {
	if false {

		//fmt.Printf("x: %v y: %v\n", deltaTx, deltaTy)
		return colour.RGB(texture.SampleRGB(c.Filename, sg))
	}

	return colour.RGB(texture.SampleFeline(c.Filename, sg))

}
