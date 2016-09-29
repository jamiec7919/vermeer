// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

#define Big		0x4330000000000000 // 2**52

// Unlike the Go package we require SSE4.

// func Floor(x float32) float32
TEXT ·Floor(SB),NOSPLIT,$0
	ROUNDSS $1, x+0(FP), X0
	MOVSS X0, ret+8(FP)
	RET

// func Ceil(x float64) float64
TEXT ·Ceil(SB),NOSPLIT,$0
	ROUNDSS $2, x+0(FP), X0
	MOVSS X0, ret+8(FP)
	RET

