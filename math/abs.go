// Copyright 2017 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// Abs returns the absolute value of x.
//
// Special cases are:
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
//
func Abs(x float32) float32 {
	return Andf(x, ^uint32(1<<31)) //xorf(x, sign_mask(x))
}

func abs(x float32) float32 {
	switch {
	case x < 0:
		return -x
	case x == 0:
		return 0 // return correctly abs(-0)
	}
	return x
}

// sign returns 1 if v is +ve or -1 if -ve
func sign(v float32) float32 {
	if v >= 0.0 {
		return 1.0
	}

	return -1.0

}
