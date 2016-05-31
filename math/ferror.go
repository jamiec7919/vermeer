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

// EpsilonFloat32 is the C++ FLT_EPSILON epsilon constant
const EpsilonFloat32 float32 = 1.19209290E-07

// MachineEpsilon32 is the smallest useful epsilon (see PBRT), same as C++ numeric_limits<float>::epsilon() * 0.5
const MachineEpsilon32 float32 = 1.19209290E-07 * 0.5

// Gamma returns the nth gamma epsilon
func Gamma(n int32) float32 {
	return (float32(n) * MachineEpsilon32) / (1 - float32(n)*MachineEpsilon32)
}

// NextFloatUp returns the next (ulp) float up from v
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

// NextFloatDown returns the next (ulp) float down from v
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
