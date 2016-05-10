// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package colour

import (
	m "github.com/jamiec7919/vermeer/math"
)

type RGB [3]float32

type RGBA [4]float32

func (c *RGB) Scale(f float32) {
	for k := range c {
		c[k] *= f
	}
}

func (c *RGB) Add(o RGB) {
	for k, v := range o {
		c[k] += v
	}
}

func (c *RGB) Mul(o RGB) {
	for k, v := range o {
		c[k] *= v
	}
}

func (c *RGB) Normalize() {

	max := m.Max(m.Max(c[0], c[1]), c[2])

	if max > 1.0 {
		for k := range c {
			c[k] *= 1 / max
		}

	}
}
