// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package colour

import (
	m "github.com/jamiec7919/vermeer/math"
)

// RGB represents a colour triple.
type RGB [3]float32

// RGBA represents a colour triple with an alpha value.
type RGBA [4]float32

// Scale scales the RGB values in-place.
func (c *RGB) Scale(f float32) {
	for k := range c {
		c[k] *= f
	}
}

// Add adds the RGB triple o to c.
func (c *RGB) Add(o RGB) {
	for k, v := range o {
		c[k] += v
	}
}

// Mul multiples the RGB triple o into c.
func (c *RGB) Mul(o RGB) {
	for k, v := range o {
		c[k] *= v
	}
}

// Normalize will scale the components of c to have a max of 1.
func (c *RGB) Normalize() {

	max := m.Max(m.Max(c[0], c[1]), c[2])

	if max > 1.0 {
		for k := range c {
			c[k] *= 1 / max
		}

	}
}
