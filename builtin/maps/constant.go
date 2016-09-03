// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
)

type Constant struct {
	C    colour.RGB
	Chan int
}

func (c *Constant) Float32(sg *core.ShaderContext) float32 {
	return c.C[c.Chan]
}

func (c *Constant) RGB(sg *core.ShaderContext) colour.RGB {
	return c.C
}
