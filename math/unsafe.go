// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import (
	"math"
	"unsafe"
)

func SignMask(v float32) uint32 {
	return sign_mask(v)
}

func sign_mask(v float32) uint32 {
	return (*(*uint32)(unsafe.Pointer(&v))) & (1 << 31) //~0x7fffffff;
}

func xorf(v float32, i uint32) float32 {
	return math.Float32frombits(*(*uint32)(unsafe.Pointer(&v)) ^ i)

}

func andf(v float32, i uint32) float32 {
	return math.Float32frombits(*(*uint32)(unsafe.Pointer(&v)) & i)

}

func Xorf(v float32, i uint32) float32 {
	return math.Float32frombits(*(*uint32)(unsafe.Pointer(&v)) ^ i)

}

func fabsf(x float32) float32 {
	return xorf(x, sign_mask(x))
}
