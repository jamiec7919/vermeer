// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import "math"

/*
	This is currently pretty terrible, it just implements conversions around the standard math library.
*/

// Sincos returns the sine of x and the cosine of x.
//
//go:nosplit
func Sincos(x float32) (float32, float32) {
	s, c := math.Sincos(float64(x))

	return float32(s), float32(c)
}

// Sin returns the sine x.
//
//go:nosplit
func Sin(x float32) float32 {
	return float32(math.Sin(float64(x)))
}

// Cos returns the cosine x.
//
//go:nosplit
func Cos(x float32) float32 {
	return float32(math.Cos(float64(x)))
}

// Acos returns the arccosine x.
//
//go:nosplit
func Acos(x float32) float32 {
	return float32(math.Acos(float64(x)))

}

// Asin returns the arccosine x.
//
//go:nosplit
func Asin(x float32) float32 {
	return float32(math.Asin(float64(x)))

}

// Atan2 returns the arctangent x/y.
//go:nosplit
func Atan2(x, y float32) float32 {
	return float32(math.Atan2(float64(x), float64(y)))

}

// Atan returns the arctangent x.
//go:nosplit
func Atan(x float32) float32 {
	return float32(math.Atan(float64(x)))

}

// Tan returns the tangent of x.
//go:nosplit
func Tan(x float32) float32 {
	return float32(math.Tan(float64(x)))
}
