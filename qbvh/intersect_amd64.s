// Copyright 2016 The Vermeer Light Tools Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// From: https://tavianator.com/fast-branchless-raybounding-box-intersections-part-2-nans/
//NOTES:  This assumes some 16-byte alignments for boxes, hits and t
// boxes is assumed to be on 16 byte alignment. Can probably combine the hits into a struct
// to ensure padding?)
// Parallel version
//func intersectBoxes(ray *Ray, boxes *[4*3*2]float32,  hits *[4]int32, t *[4]float32) 
TEXT ·intersectBoxes(SB),NOSPLIT,$0
    MOVQ    boxes+8(FP),DX
    MOVQ    ray+0(FP),AX


    MOVSS   (AX),X3  // X3 = ray.O[0]
    SHUFPS  $0x00,X3,X3  // broadcast
    MOVSS   24(AX),X4  // X4 = ray.Dinv[0]
    SHUFPS  $0x00,X4,X4  // broadcast

    LEAQ    (DX),BP    // *boxes pointer)
//    LEAQ    (BP)(CX*4),BP  // *boxes->BP

    MOVAPS   (BP),X2  // X2 = boxmin.x
    SUBPS   X3,X2  // X2 = boxmin.x - ray.O[0]
    MULPS   X4,X2  // X2 = t1 = (boxmin.x - ray.O[0]) * ray.Dinv[0]

    ADDQ    $48,BP     // boxmax.x pointer 
    MOVAPS   (BP),X6  // X6 = boxmax.x
    SUBPS   X3,X6  // X6 = boxmax.x - ray.O[0]
    MULPS   X4,X6  // X6 = (boxmax.x - ray.O[0]) * ray.Dinv[0]

//    MOVAPS   X0,X6
    MOVAPS   X6,X7  // X7 = X6 = t2
    MINPS   X2,X6   // X6 = min(t1,t2)
    MAXPS   X2,X7   // X7 = max(t1,t2)

    MOVSS   4(AX),X3  // X3 = ray.O[1]
    SHUFPS  $0x00,X3,X3  // broadcast
    MOVSS   28(AX),X4  // X4 = ray.Dinv[1]
    SHUFPS  $0x00,X4,X4  // broadcast

    SUBQ    $32,BP
    MOVAPS   (BP),X2  // X2 = boxmin.y
    SUBPS   X3,X2  // X2 = boxmin.y - ray.O[1]
    MULPS   X4,X2  // X2 = t2 = (boxmin.y - ray.O[1]) * ray.Dinv[1]

    ADDQ    $48,BP     // boxmax.y pointer 
    MOVAPS   (BP),X0  // X0 = boxmax.y
    SUBPS   X3,X0  // X0 = boxmax.y - ray.O[1]
    MULPS   X4,X0  // X0 = t2 = (boxmax.y - ray.O[1]) * ray.Dinv[1]

    MOVAPS   X0,X1   // X1 = t2
    MINPS   X2,X1   // X1 = min(t1,t2)
    MAXPS   X2,X0   // X0 = max(t1,t2)
    MAXPS   X1,X6   // tmin = max(tmin,min(t1,t2)) 
    MINPS   X0,X7   // tmax = min(tmax,max(t1,t2)

    MOVSS   8(AX),X3  // X3 = ray.O[2]
    SHUFPS  $0x00,X3,X3  // broadcast
    MOVSS   32(AX),X4  // X4 = ray.Dinv[2]
    SHUFPS  $0x00,X4,X4  // broadcast

    SUBQ     $32,BP

    MOVAPS   (BP),X2  // X2 = boxmin.z
    SUBPS   X3,X2  // X2 = boxmin.z - ray.O[2]
    MULPS   X4,X2  // X2 = t2 = (boxmin.z - ray.O[2]) * ray.Dinv[2]

    ADDQ    $48,BP     // boxmax.z pointer 
    MOVAPS   (BP),X0  // X0 = boxmax.z
    SUBPS   X3,X0  // X0 = boxmax.z - ray.O[2]
    MULPS   X4,X0  // X0 = t2 = (boxmax.z - ray.O[2]) * ray.Dinv[2]

    MOVAPS   X0,X1   // X1 = t2
    MINPS   X2,X1   // X1 = min(t1,t2)
    MAXPS   X2,X0   // X0 = max(t1,t2)
    MAXPS   X1,X6   // tmin = max(tmin,min(t1,t2)) 
    MINPS   X0,X7   // tmax = min(tmax,max(t1,t2)

    // X6 = tmin,  X7 = tmax 
    //MOVSS $f32.00000000+0(SB),X0
    //MOVSS   $0.00000000,X0     // ?? constants?
    //SHUFPS  $0x00,X0,X0
    XORPS   X0,X0    // X0 = 0.0
    MAXPS   X6,X0   // X0 = tNear = MAX(0,tmin)

    MOVQ    t+24(FP),DX    // *t [4]float32 pointer
    LEAQ    (DX),BP
    MOVAPS   X0,(BP)

    CMPPS  X7,X0,$0x2    // X7 >= X0  tmax >= tNear

    MOVQ    hits+16(FP),DX    // *hits [4]int32 pointer
    LEAQ    (DX),BP
    MOVAPS   X0,(BP)

    RET
