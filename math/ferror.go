// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import (
	"math"
)

/*
	Floating point error computation, some (all) of this based on FP error handling from PBRT.
*/

const FLT_EPSILON float32 = 1.19209290E-07            // C++ epsilon
const MachineEpsilon32 float32 = 1.19209290E-07 * 0.5 //numeric_limits<float>::epsilon() * 0.5

func Gamma(n int32) float32 {
	return (float32(n) * MachineEpsilon32) / (1 - float32(n)*MachineEpsilon32)
}

/*
Already exist in math library
func Float32ToBits(v float32) uint32 {
	return *(*uint32)(unsafe.Pointer(&v))
}

func BitsToFloat32(v uint32) float32 {
	return *(*float32)(unsafe.Pointer(&v))
}
*/

func NextFloatUp(v float32) float32 {
	if IsInf(v) && v > 0 {
		return v
	}

	if v == -0.0 {
		v = 0.0
	}

	ui := math.Float32bits(v)

	if v >= 0.0 {
		ui++
	} else {
		ui--
	}

	return math.Float32frombits(ui)
}

func NextFloatDown(v float32) float32 {
	if IsInf(v) && v < 0 {
		return v
	}

	if v == -0.0 {
		v = 0.0
	}

	ui := math.Float32bits(v)

	if v >= 0.0 {
		ui--
	} else {
		ui++
	}

	return math.Float32frombits(ui)

}
