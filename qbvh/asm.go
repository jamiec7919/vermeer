// +build asm

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
)

func main() {
	Package("github.com/jamiec7919/vermeer/qbvh")
	Implement("IntersectBoxes")

	ray := Dereference(Param("ray"))

	Px := Load(ray.Field("P").Field("X"), XMM())
	SHUFPS(U8(0), Px, Px)
	Dinvx := Load(ray.Field("Dinv").Field("X"), XMM())
	SHUFPS(U8(0), Dinvx, Dinvx)

	box := Load(Param("boxes"), GP64())
	t1 := XMM()
	MOVAPS(Mem{Base: box}, t1)
	SUBPS(Px, t1)
	MULPS(Dinvx, t1)

	t2 := XMM()
	MOVAPS(Mem{Base: box, Disp: 48}, t2) // Could we get the mem via the type?
	SUBPS(Px, t2)
	MULPS(Dinvx, t2)

	tmin := XMM()
	tmax := XMM()
	MOVAPS(t1, tmin)
	MOVAPS(t1, tmax)
	MINPS(t2, tmin)
	MAXPS(t2, tmax)

	Py := Load(ray.Field("P").Field("Y"), XMM())
	SHUFPS(U8(0), Py, Py)
	Dinvy := Load(ray.Field("Dinv").Field("Y"), XMM())
	SHUFPS(U8(0), Dinvy, Dinvy)

	MOVAPS(Mem{Base: box, Disp: 16}, t1)
	SUBPS(Py, t1)
	MULPS(Dinvy, t1)

	MOVAPS(Mem{Base: box, Disp: 16 + 48}, t2) // Could we get the mem via the type?
	SUBPS(Py, t2)
	MULPS(Dinvy, t2)

	tmp0 := XMM()

	MOVAPS(t1, tmp0)
	MINPS(t2, t1)
	MAXPS(tmp0, t2)

	MINPS(t2, tmax)
	MAXPS(t1, tmin)

	Pz := Load(Dereference(Param("ray")).Field("P").Field("Z"), XMM())
	SHUFPS(U8(0), Pz, Pz)
	Dinvz := Load(Dereference(Param("ray")).Field("Dinv").Field("Z"), XMM())
	SHUFPS(U8(0), Dinvz, Dinvz)

	MOVAPS(Mem{Base: box, Disp: 32}, t1)
	SUBPS(Pz, t1)
	MULPS(Dinvz, t1)

	MOVAPS(Mem{Base: box, Disp: 32 + 48}, t2) // Could we get the mem via the type?
	SUBPS(Pz, t2)
	MULPS(Dinvz, t2)

	//tmp0 := XMM()

	MOVAPS(t1, tmp0)
	MINPS(t2, t1)
	MAXPS(tmp0, t2)

	MINPS(t2, tmax)
	MAXPS(t1, tmin)

	tNear := XMM()
	XORPS(tNear, tNear)
	MAXPS(tmin, tNear) // tNear = max(tmin,0)

	t := Load(Param("t"), GP64())
	MOVAPS(tNear, Mem{Base: t})

	hits := Load(Param("hits"), GP64())
	CMPPS(tmax, tNear, U8(0x2)) // tmax >= tNear?
	MOVAPS(tNear, Mem{Base: hits})

	RET()

	Generate()
}
