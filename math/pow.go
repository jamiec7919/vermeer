// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import (
	m "math"
)

// Log2 returns log2(x)
func Log2(x float32) float32 {
	return float32(m.Log2(float64(x)))
}

// Pow returns x^y
func Pow(x, y float32) float32 {
	return float32(m.Pow(float64(x), float64(y)))
}
