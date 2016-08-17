// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import (
	"math"
	"unsafe"
)

// Can we use andnps on amd64?

// SignMask returns the bits of the float32 v as a uint32 with everything masked off except the sign bit.
func SignMask(v float32) uint32 {
	return signMask(v)
}

func signMask(v float32) uint32 {
	return (*(*uint32)(unsafe.Pointer(&v))) & (1 << 31) //~0x7fffffff;
}

func xorf(v float32, i uint32) float32 {
	return math.Float32frombits(*(*uint32)(unsafe.Pointer(&v)) ^ i)

}

func andf(v float32, i uint32) float32 {
	return math.Float32frombits(*(*uint32)(unsafe.Pointer(&v)) & i)

}

// Xorf returns the bitwise xor of v with i.
func Xorf(v float32, i uint32) float32 {
	return math.Float32frombits(*(*uint32)(unsafe.Pointer(&v)) ^ i)

}

func fabsf(x float32) float32 {
	return xorf(x, signMask(x))
}
