// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

#define Half 0.5
#define OnePointFive 1.5

// func ·Vec3Normalize(a Vec3) (b Vec3)
TEXT ·Vec3Normalize(SB),NOSPLIT,$0
	MOVSS a+0(FP), X0
	MOVSS a+4(FP), X1
	MOVSS a+8(FP), X2
	MULSS X0,X0
	MULSS X1,X1
	MULSS X2,X2
	ADDSS X0,X1 
	ADDSS X2,X1
	RSQRTSS X1,X0

	// X0 is guess rsqrt. X1 is guess^2  Use Newton-Raphson to improve.
	MOVSS $Half,X2
	MULSS X1,X2      // X2 = guess^2/2
    MOVSS X0,X3      // save guess
    MULSS X0,X0
    MULSS X2,X0      // 
    MOVSS $OnePointFive,X4
    SUBSS X0,X4
    MOVSS X4,X0
    MULSS X3,X0  // X0 now holds corrected guess

	MOVSS a+0(FP), X1
    MULSS X0,X1 
    MOVSS X1,v+16(FP)
	MOVSS a+4(FP), X1
    MULSS X0,X1 
    MOVSS X1,v+20(FP)
	MOVSS a+8(FP), X1
    MULSS X0,X1 
    MOVSS X1,v+24(FP)
	RET

// func ·Vec3Length(a Vec3) float32
TEXT ·Vec3Length(SB),NOSPLIT,$0
	MOVSS a+0(FP), X0
	MOVSS a+4(FP), X1
	MOVSS a+8(FP), X2
	MULSS X0,X0
	MULSS X1,X1
	MULSS X2,X2
	ADDSS X0,X1 
	ADDSS X2,X1
	SQRTSS X1,X0
    MOVSS X0,ret+16(FP)
    RET

    