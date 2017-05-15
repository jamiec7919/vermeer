// Copyright 2017 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import (
	"math"
)

// InfPos is +ve infinity in 32 bits.
var InfPos = math.Float32frombits(uint32(0xff) << 23)

// InfNeg is -ve infinity in 32 bits.
var InfNeg = math.Float32frombits(uint32(0x1ff) << 23)

// Inf returns either positive or negative infinity based on sign.
func Inf(sign int) float32 {
	if sign > 0 {
		return InfPos
	}

	return InfNeg

}

// IsInf returns true if the value represented by v is infinity.
func IsInf(v float32) bool {
	return v > math.MaxFloat32 || v < -math.MaxFloat32
}
