// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// func ·Max(x, y float32) float32
TEXT ·Max(SB),NOSPLIT,$0
	MOVSS x+0(FP), X0
	MOVSS y+4(FP), X1
	MAXSS X1, X0
	MOVSS X0, ret+8(FP)
	RET

// func Min(x, y float64) float64
TEXT ·Min(SB),NOSPLIT,$0
	MOVSS x+0(FP), X0
	MOVSS y+4(FP), X1
	MINSS X1, X0
	MOVSS X0, ret+8(FP)
	RET
