// Code generated by command: go run asm.go -out asm_amd64.s. DO NOT EDIT.

// func IntersectBoxes(ray *core.Ray, boxes *[24]float32, hits *[4]int32, t *[4]float32)
TEXT ·IntersectBoxes(SB), $0-32
	MOVQ   ray+0(FP), AX
	MOVSS  (AX), X0
	SHUFPS $0x00, X0, X0
	MOVSS  24(AX), X1
	SHUFPS $0x00, X1, X1
	MOVQ   boxes+8(FP), CX
	MOVAPS (CX), X2
	SUBPS  X0, X2
	MULPS  X1, X2
	MOVAPS 48(CX), X3
	SUBPS  X0, X3
	MULPS  X1, X3
	MOVAPS X2, X0
	MOVAPS X2, X1
	MINPS  X3, X0
	MAXPS  X3, X1
	MOVSS  4(AX), X5
	SHUFPS $0x00, X5, X5
	MOVSS  28(AX), X6
	SHUFPS $0x00, X6, X6
	MOVAPS 16(CX), X2
	SUBPS  X5, X2
	MULPS  X6, X2
	MOVAPS 64(CX), X3
	SUBPS  X5, X3
	MULPS  X6, X3
	MOVAPS X2, X5
	MINPS  X3, X2
	MAXPS  X5, X3
	MINPS  X3, X1
	MAXPS  X2, X0
	MOVQ   ray+0(FP), AX
	MOVSS  8(AX), X5
	SHUFPS $0x00, X5, X5
	MOVQ   ray+0(FP), AX
	MOVSS  32(AX), X6
	SHUFPS $0x00, X6, X6
	MOVAPS 32(CX), X2
	SUBPS  X5, X2
	MULPS  X6, X2
	MOVAPS 80(CX), X3
	SUBPS  X5, X3
	MULPS  X6, X3
	MOVAPS X2, X5
	MINPS  X3, X2
	MAXPS  X5, X3
	MINPS  X3, X1
	MAXPS  X2, X0
	XORPS  X4, X4
	MAXPS  X0, X4
	MOVQ   t+24(FP), AX
	MOVAPS X4, (AX)
	MOVQ   hits+16(FP), AX
	CMPPS  X1, X4, $0x02
	MOVAPS X4, (AX)
	RET
