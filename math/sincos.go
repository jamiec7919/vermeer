// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import "math"

/*
	This is currently pretty terrible, it just implements conversions around the standard math library.
*/

// Sin returns the sine x.
//
func Sin(x float32) float32 {
	return float32(math.Sin(float64(x)))
}

// Cos returns the cosine x.
//
func Cos(x float32) float32 {
	return float32(math.Cos(float64(x)))
}

func Acos(x float32) float32 {
	return float32(math.Acos(float64(x)))

}

func Atan2(x, y float32) float32 {
	return float32(math.Atan2(float64(x), float64(y)))

}

func Tan(x float32) float32 {
	return float32(math.Tan(float64(x)))
}
