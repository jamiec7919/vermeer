// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import "math"

var InfPos = math.Float32frombits(uint32(0xff) << 23)
var InfNeg = math.Float32frombits(uint32(0x1ff) << 23)

func Inf(sign int) float32 {
	if sign > 0 {
		return InfPos
	} else {
		return InfNeg
	}
}

func IsInf(v float32) bool {
	return v > math.MaxFloat32 || v < -math.MaxFloat32
}
