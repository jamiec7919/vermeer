/*

stacktop = 0   SI
stack[stacktop].T = ray.Tclosest   
stack[stacktop].Node = 0

for stacktop >= 0 {

	check if ray.Tclosest < stack[stacktop].T, if so then pop it and continue.

	node = stack[stacktop].Node   BX

	if node < 0 
      we're in leaf, grab base and count and call traceTris
    else
      intersect ray with all 4 boxes of node 

      for each child
        if ray hits box 
          stack[stacktop].T = boxintersection of child.T
          stack[stacktop].Node = child
          stacktop++

          check this isn't an empty leaf (i.e. child==-1), if so then stacktop--

    stacktop--    <- make sure next iteration will be at top
}

stack is at ray+32, node size is 8 bytes  = { Node int32,  T float32}
nodes is a slice so pointer to nodes is at nodes+16
*/

// Copyright 2016 The Vermeer Light Tools Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

//func traceRayAccel_asm(ray *core.RayData, mesh *Mesh, nodes []qbvh.Node)
TEXT ·traceRayAccel_asm(SB),NOSPLIT,$32-40

   MOVQ    nodes_base+16(FP),CX     // Assume pointer to slice memory is first element of header??
   MOVQ    mesh+8(FP),R10         // R10 is mesh pointer
   // CX is pointer to nodes 
   MOVQ		ray+0(FP),BX       // BX is pointer to base of ray 

   XORQ   AX,AX 

  //SUBQ   $32,SP    // arg space
  MOVQ   BX,(SP)   // ray as first arg
  // ax has node, BX&DX is free to use
  MOVQ   AX,DX 
  ANDL	$134217727, DX
  SARL	$4, DX
  MOVLQSX	DX, DX
  ANDL	$15, AX
  INCL	AX
  MOVLQSX	AX, AX
  MOVQ   DX,8(SP)
  MOVQ   AX,16(SP)
  MOVQ   R10,24(SP)
  CALL   ·meshTraceTris_new(SB)   // ray*, base,count int, mesh*
  //ADDQ   $32,SP

   RET 
/*
//func traceRayAccel_asm(ray *core.RayData, mesh *Mesh, nodes []qbvh.Node)
TEXT ·_traceRayAccel_asm(SB),NOSPLIT,$0

   MOVQ    node_ptr+16(FP),CX     // Assume pointer to slice memory is first element of header??
   MOVQ    mesh+8(FP),R10         // R10 is mesh pointer
   // CX is pointer to nodes 
   MOVQ		ray+0(FP),BX       // BX is pointer to base of ray 
   LEAQ     8(BX),DX           // (BP) is pointer to base of stack
   //MOVQ     BP,DX            // DX is pointer to base of stack 
   XORPS    X0,X0  				// X0 = 0
   MOVSS    X0,(DX)            // stack[stacktop].Node = 0
   LEAQ     548(BX),BP         // Tclosest
   MOVSS     (BP),X0            // X0 = Tclosest
   MOVSS     X0,4(DX)           // stack[stacktop].T = ray.Tclosest

	SUBQ 	$32,SP 

   XORQ  	SI,SI              // stacktop

stackloop:
   LEAQ     548(BX),BP         // Tclosest
   MOVSS     (BP),X0            // X0 = Tclosest

  MOVSS     4(DX)(SI*8),X1
  UCOMISS   X0,X1             // ray.Tclosest < stack[stacktop].T  
  JNC       dontpop                  // carry flag if op1 < op2
  DECQ       SI 
  CMPQ       SI,$0
  JGE       stackloop 
  JMP       endloop

dontpop:
  MOVL   (DX)(SI*8),AX        // AX = stack[stacktop].Node 

  CMPQ   AX,$0                //  if node < 0
  JGE    isnode               // not a leaf
// leaf:
// save:  DX,SI,BX 
  MOVQ   DX,(SP)
  MOVQ   SI,8(SP)
  MOVQ   BX,16(SP)

  SUBQ   $32,SP    // arg space
  MOVQ   BX,(SP)   // ray as first arg
  // ax has node, BX&DX is free to use
  MOVQ   AX,DX 
  ANDL	$134217727, DX
  SARL	$4, DX
  MOVLQSX	DX, DX
  ANDL	$15, AX
  INCL	AX
  MOVLQSX	AX, AX
  MOVQ   DX,8(SP)
  MOVQ   AX,16(SP)
  MOVQ   R10,24(SP)
  CALL   ·meshTraceTris(SB)   // ray*, base,count int, mesh*
  ADDQ   $32,SP
  MOVQ   mesh+8(FP),R10
  MOVQ   (SP),DX 
  MOVQ   8(SP),SI 
  MOVQ   16(SP),BX
  JMP    stackloop
isnode:
// Do intersection for node in AX
// CX is pointer to nodes 
    
	SHLQ  $7,AX 		// AX * 128 is offset to node   
    LEAQ  (CX)(AX*1),BP   // (BP) is boxes data 

    MOVSS   512(BX),X3  // X3 = ray.O[0]
    SHUFPS  $0x00,X3,X3  // broadcast
    MOVSS   512+24(BX),X4  // X4 = ray.Dinv[0]
    SHUFPS  $0x00,X4,X4  // broadcast

    MOVAPS   (BP),X2  // X2 = boxmin.x
    SUBPS   X3,X2  // X2 = boxmin.x - ray.O[0]
    MULPS   X4,X2  // X2 = t1 = (boxmin.x - ray.O[0]) * ray.Dinv[0]

    ADDQ    $48,BP     // boxmax.x pointer 
    MOVAPS   (BP),X6  // X6 = boxmax.x
    SUBPS   X3,X6  // X6 = boxmax.x - ray.O[0]
    MULPS   X4,X6  // X6 = (boxmax.x - ray.O[0]) * ray.Dinv[0]

    MOVAPS   X6,X7  // X7 = X6 = t2
    MINPS   X2,X6   // X6 = min(t1,t2)
    MAXPS   X2,X7   // X7 = max(t1,t2)

    MOVSS   512+4(BX),X3  // X3 = ray.O[1]
    SHUFPS  $0x00,X3,X3  // broadcast
    MOVSS   512+28(BX),X4  // X4 = ray.Dinv[1]
    SHUFPS  $0x00,X4,X4  // broadcast

    SUBQ    $32,BP     // BP = boxin.y pointer 

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

    MOVSS   512+8(BX),X3  // X3 = ray.O[2]
    SHUFPS  $0x00,X3,X3  // broadcast
    MOVSS   512+32(BX),X4  // X4 = ray.Dinv[2]
    SHUFPS  $0x00,X4,X4  // broadcast

    SUBQ    $32,BP     // BP = boxmin.z pointer 

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

    XORPS   X0,X0    // X0 = 0.0
    MAXPS   X6,X0   // X0 = tNear = MAX(0,tmin)

   // X0 now contains tNear for all children
   MOVAPS   X0,(SP)

    CMPPS  X7,X0,$0x2    // X7 >= X0  tmax >= tNear

  // X0 now has hit mask 
   MOVAPS  X0,16(SP)

  // tNear + hits are on stack  

  MOVQ    $3,DI   // 3,2,1,0
child:
  MOVL    16(SP)(DI*4),AX
  CMPL    AX,$0            // is hit?
  JE     hit
  DECQ   DI 
  CMPQ   DI,$0
  JGE    child 
  JMP    endchild
hit:
  // child is at BP + 16+12 + (dx*4)
  MOVL   16+12(BP)(DI*4),AX    // AX is child index
  MOVL   AX,(DX)(SI*8)         // stack[stacktop].Node = child[k]
  MOVSS  (SP)(DI*4),X0 
  MOVSS  X0,4(DX)(SI*8)        // stack[stacktop].T = tNearest[k]
  INCQ   SI                    // stacktop++ 

  CMPQ   AX,$-1
  JNE    notemptyleaf
  DECQ   SI 

notemptyleaf: 
  DECQ   DI          // loop on child,  k++ 
  CMPQ   DI,$0
  JGE    child 
endchild:
  DECQ   SI 
  CMPQ   SI,$0
  JL    endloop
  JMP    stackloop

endloop:
  ADDQ   $32,SP
  RET   

*/