// Copyright 2016 The Vermeer Light Tools Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

//func traceRayAccel_asm(f []FaceGeom,nodes []qbvh.Node,ray *Ray)
/*TEXT ·traceRayAccel_asm(SB),NOSPLIT,$0
    // f is slice of FaceGeoms, load:
    MOVQ    f_ptr+16(FP),DX
    MOVQ    node_ptr+40(FP),CX

    SUBQ    $24,SP  // allocate some stack space

    LEAQ ·traverseStack(SB), BX

    MOVQ  $0,AX  // stackTop
    //  ray.transformRay.Tclosest is at R6+(64+9*4)= (R6+100)
    MOVQ    ray+48(FP),BP
    LEAQ  100(BP),BP
    XORPS X0,X0
    MOVSS X0,(BX)   // traverseStack[stackTop].node = 0
    MOVSS (BP),X0
    MOVSS  X0,4(BX)  // traverseStack[stackTop].Tclosest = ray.transformRay.Tclosest

    RET 
*/
//NOTES:  This assumes some 16-byte alignments (although removed that from ht&tnear.
// node is assumed to be on 16 byte alignment. Can probably combine the hits into a struct
// to ensure padding?)
// Parallel version
//func rayNodeIntersectAllASM(ray *Ray, node *qbvh.Node,  hit *[4]int32, tNear *[4]float32) 
TEXT ·rayNodeIntersectAllASM(SB),NOSPLIT,$0
    MOVQ    node+8(FP),DX
    MOVQ    ray+0(FP),AX


    MOVSS   (AX),X3  // X3 = ray.O[0]
    SHUFPS  $0x00,X3,X3  // broadcast
    MOVSS   24(AX),X4  // X4 = ray.Dinv[0]
    SHUFPS  $0x00,X4,X4  // broadcast

    LEAQ    (DX),BP    // *node -> BP  (same as *boxes pointer)
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

    MOVQ    tNear+24(FP),DX    // *tNear [4]float32 pointer
    LEAQ    (DX),BP
    MOVAPS   X0,(BP)

    CMPPS  X7,X0,$0x2    // X7 >= X0  tmax >= tNear

    MOVQ    hits+16(FP),DX    // *tNear [4]float32 pointer
    LEAQ    (DX),BP
    MOVAPS   X0,(BP)

    RET

//NOTES:  This assumes some 16-byte alignments (although removed that from ht&tnear.
// node is assumed to be on 16 byte alignment. Can probably combine the hits into a struct
// to ensure padding?)
// Parallel version
//func rayNodeIntersectAll_asm(ray *Ray, node *qbvh.Node,  hit *[4]int32, tNear *[4]float32) 
TEXT ·rayNodeIntersectAll2_asm(SB),NOSPLIT,$0
    MOVQ    node+8(FP),DX
    MOVQ    ray+0(FP),AX

    MOVQ    DX,SI 

    MOVSS   (AX),X3  // X3 = ray.O[0]
    SHUFPS  $0x00,X3,X3  // broadcast
    MOVSS   24(AX),X4  // X4 = ray.Dinv[0]
    SHUFPS  $0x00,X4,X4  // broadcast

    LEAQ    (DX),BP    // *node -> BP  (same as *boxes pointer)
//    LEAQ    (BP)(CX*4),BP  // *boxes->BP

    MOVAPS   (BP),X2  // X2 = boxmin.x
    SUBPS   X3,X2  // X2 = boxmin.x - ray.O[0]
    MULPS   X4,X2  // X2 = t1 = (boxmin.x - ray.O[0]) * ray.Dinv[0]

    ADDQ    $48,DX     // boxmax.x pointer 
    LEAQ    (DX),BP    // *node -> BP  (same as *boxes pointer)
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

    MOVQ    SI,DX
    ADDQ    $16,DX     // boxmin.y pointer 
    LEAQ    (DX),BP    // *node -> BP  (same as *boxes pointer)

    MOVAPS   (BP),X2  // X2 = boxmin.y
    SUBPS   X3,X2  // X2 = boxmin.y - ray.O[1]
    MULPS   X4,X2  // X2 = t2 = (boxmin.y - ray.O[1]) * ray.Dinv[1]

    ADDQ    $48,DX     // boxmax.y pointer 
    LEAQ    (DX),BP    // *node -> BP  (same as *boxes pointer)
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

    MOVQ    SI,DX
    ADDQ    $32,DX     // boxmin.z pointer 
    LEAQ    (DX),BP    // *node -> BP  (same as *boxes pointer)

    MOVAPS   (BP),X2  // X2 = boxmin.z
    SUBPS   X3,X2  // X2 = boxmin.z - ray.O[2]
    MULPS   X4,X2  // X2 = t2 = (boxmin.z - ray.O[2]) * ray.Dinv[2]

    ADDQ    $48,DX     // boxmax.z pointer 
    LEAQ    (DX),BP    // *node -> BP  (same as *boxes pointer)
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

    MOVQ    tNear+24(FP),DX    // *tNear [4]float32 pointer
    LEAQ    (DX),BP
    MOVAPS   X0,(BP)

    CMPPS  X7,X0,$0x2    // X7 >= X0  tmax >= tNear

    MOVQ    hits+16(FP),DX    // *tNear [4]float32 pointer
    LEAQ    (DX),BP
    MOVAPS   X0,(BP)

    RET

// Parallel version
//func rayNodeIntersectAll_asm(ray *Ray, node *qbvh.Node,  hit *[4]int32, tNear *[4]float32) 
TEXT ·rayNodeIntersectAll_asm2(SB),NOSPLIT,$0
    MOVQ    node+8(FP),DX
    MOVQ    ray+0(FP),AX

    MOVQ    DX,SI 

    MOVSS   (AX),X3  // X3 = ray.O[0]
    SHUFPS  $0x00,X3,X3  // broadcast
    MOVSS   24(AX),X4  // X4 = ray.Dinv[0]
    SHUFPS  $0x00,X4,X4  // broadcast

    LEAQ    (DX),BP    // *node -> BP  (same as *boxes pointer)
//    LEAQ    (BP)(CX*4),BP  // *boxes->BP

    MOVAPS   (BP),X2  // X2 = boxmin.x
    SUBPS   X3,X2  // X2 = boxmin.x - ray.O[0]
    MULPS   X4,X2  // X2 = (boxmin.x - ray.O[0]) * ray.Dinv[0]

    ADDQ    $48,DX     // boxmax.x pointer 
    LEAQ    (DX),BP    // *node -> BP  (same as *boxes pointer)
    MOVAPS   (BP),X0  // X0 = boxmax.x
    SUBPS   X3,X0  // X0 = boxmax.x - ray.O[0]
    MULPS   X4,X0  // X0 = (boxmax.x - ray.O[0]) * ray.Dinv[0]

    MOVAPS   X0,X6
    MOVAPS   X0,X7 
    MINPS   X2,X6   // X6 = min(t1,t2)
    MAXPS   X2,X7   // X7 = max(t1,t2)

    MOVSS   4(AX),X3  // X3 = ray.O[1]
    SHUFPS  $0x00,X3,X3  // broadcast
    MOVSS   28(AX),X4  // X4 = ray.Dinv[1]
    SHUFPS  $0x00,X4,X4  // broadcast

    MOVQ    SI,DX
    ADDQ    $16,DX     // boxmin.y pointer 
    LEAQ    (DX),BP    // *node -> BP  (same as *boxes pointer)

    MOVAPS   (BP),X2  // X2 = boxmin.y
    SUBPS   X3,X2  // X2 = boxmin.y - ray.O[1]
    MULPS   X4,X2  // X2 = t2 = (boxmin.y - ray.O[1]) * ray.Dinv[1]

    ADDQ    $48,DX     // boxmax.y pointer 
    LEAQ    (DX),BP    // *node -> BP  (same as *boxes pointer)
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

    MOVQ    SI,DX
    ADDQ    $32,DX     // boxmin.z pointer 
    LEAQ    (DX),BP    // *node -> BP  (same as *boxes pointer)

    MOVAPS   (BP),X2  // X2 = boxmin.z
    SUBPS   X3,X2  // X2 = boxmin.z - ray.O[2]
    MULPS   X4,X2  // X2 = t2 = (boxmin.z - ray.O[2]) * ray.Dinv[2]

    ADDQ    $48,DX     // boxmax.z pointer 
    LEAQ    (DX),BP    // *node -> BP  (same as *boxes pointer)
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
    MOVSS   $0.00000000,X0     // ?? constants?
    SHUFPS  $0x00,X0,X0
    MAXPS   X6,X0   // X0 = tNear = MAX(0,tmin)

    MOVQ    tNear+24(FP),DX    // *tNear [4]float32 pointer
    LEAQ    (DX),BP
    MOVAPS   X0,(BP)

    CMPPS  X7,X0,$0x2    // X7 >= X0  tmax >= tNear

    MOVQ    hits+16(FP),DX    // *tNear [4]float32 pointer
    LEAQ    (DX),BP
    MOVAPS   X0,(BP)

    RET


// X6 = tmin
// X7 = tmax
// func rayNodeIntersect2(ray *Ray, node *qbvh.Node, idx int) (bool, float32)
TEXT ·rayNodeIntersect2(SB),NOSPLIT,$0
    MOVQ	idx+16(FP),CX
    MOVQ	node+8(FP),DX
    MOVQ	ray+0(FP),AX

    MOVQ    CX,SI  // Save CX      

	MOVSS	(AX),X3  // X3 = ray.O[0]
	MOVSS	24(AX),X4  // X4 = ray.Dinv[0]
    LEAQ	(DX),BP    // *node -> BP
    LEAQ	(BP)(CX*4),BP  // *boxes->BP
    MOVSS	(BP),X2  // X2 = boxmin.x
	SUBSS	X3,X2    // X2 = box.x - ray.O[0]
	MULSS	X4,X2   // X2 = (box.x - ray.O[0]) * ray.Dinv[0]
    ADDQ    $12,CX  // CX = next boxmax.x
    LEAQ	(DX),BP    // *node -> BP
    LEAQ	(BP)(CX*4),BP  // *boxes->BP
    MOVSS	(BP),X0  // X0 = boxmax.x
	// MOVSS	(AX),X1  // X1 = ray.O[0]
	SUBSS	X3,X0    // X0 = box.x - ray.O[0]
	// MOVSS	12(AX),X1  // X1 = ray.Dinv[0]
	MULSS	X4,X0   // X0 = (box.x - ray.O[0]) * ray.Dinv[0]
	MOVSS   X0,X6
	MOVSS   X0,X7 
	MINSS   X2,X6   // X6 = min(t1,t2)
	MAXSS   X2,X7   // X7 = max(t1,t2)

	MOVSS	4(AX),X3  // X3 = ray.O[1]
	MOVSS	28(AX),X4  // X4 = ray.Dinv[1]
	MOVQ 	SI,CX   // Restore CX
    ADDQ    $4,CX  // CX = next boxmin.y
    LEAQ	(DX),BP    // *node -> BP
    LEAQ	(BP)(CX*4),BP  // *boxes->BP
    MOVSS	(BP),X2  // X2 = boxmin.y
	SUBSS	X3,X2    // X2 = box.x - ray.O[1]
	MULSS	X4,X2   // X2 = t1 =  (box.x - ray.O[1]) * ray.Dinv[1]
	ADDQ    $12,CX
    LEAQ	(DX),BP    // *node -> BP
    LEAQ	(BP)(CX*4),BP  // *boxes->BP
    MOVSS	(BP),X0  // X0 = boxmax.x
	//MOVSS	4(AX),X1  // X1 = ray.O[1]
	SUBSS	X3,X0    // X0 = box.x - ray.O[1]
	//MOVSS	16(AX),X1  // X1 = ray.Dinv[1]
	MULSS	X4,X0   // X0 = t2 = (box.x - ray.O[1]) * ray.Dinv[1]
    MOVSS   X0,X1   // X1 = t2
    MINSS   X2,X1   // X1 = min(t1,t2)
    MAXSS   X2,X0   // X0 = max(t1,t2)
    MAXSS   X1,X6   // tmin = max(tmin,min(t1,t2)) 
    MINSS   X0,X7   // tmax = min(tmax,max(t1,t2)

	MOVSS	8(AX),X3  // X3 = ray.O[2]
	MOVSS	32(AX),X4  // X4 = ray.Dinv[2]
	MOVQ 	SI,CX   // Restore CX
    ADDQ    $8,CX  // CX = next boxmin.y
    LEAQ	(DX),BP    // *node -> BP
    LEAQ	(BP)(CX*4),BP  // *boxes->BP
    MOVSS	(BP),X2  // X2 = boxmin.y
	SUBSS	X3,X2    // X2 = box.x - ray.O[2]
	MULSS	X4,X2   // X2 = t1 =  (box.x - ray.O[2]) * ray.Dinv[2]
	ADDQ    $12,CX
    LEAQ	(DX),BP    // *node -> BP
    LEAQ	(BP)(CX*4),BP  // *boxes->BP
    MOVSS	(BP),X0  // X0 = boxmax.x
	//MOVSS	8(AX),X1  // X1 = ray.O[2]
	SUBSS	X3,X0    // X0 = box.x - ray.O[2]
	//MOVSS	20(AX),X1  // X1 = ray.Dinv[2]
	MULSS	X4,X0   // X0 = t2 = (box.x - ray.O[2]) * ray.Dinv[2]
    MOVSS   X0,X1   // X1 = t2
    MINSS   X2,X1   // X1 = min(t1,t2)
    MAXSS   X2,X0   // X0 = max(t1,t2)
    MAXSS   X1,X6   // tmin = max(tmin,min(t1,t2)) 
    MINSS   X0,X7   // tmax = min(tmax,max(t1,t2)

    // X6 = tmin,  X7 = tmax 
    //MOVSS	$f32.00000000+0(SB),X0
    MOVSS	$0.00000000,X0     // ?? constants?
    MAXSS   X6,X0   // X0 = tNear = MAX(0,tmin)

 
    UCOMISS	X0,X7   // tmax >= tNear
    JCC tmaxLT
    // tmax < tNear
   MOVSS  X0,tNear+28(FP)  // tNear returned regardless
    MOVB  $0,b+24(FP)
    RET
tmaxLT:
   MOVSS  X0,tNear+28(FP)  // tNear returned regardless
   MOVB  $1,b+24(FP)
 
	RET
