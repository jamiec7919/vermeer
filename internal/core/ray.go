// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	//"log"
	//"unsafe"
	"errors"
	"github.com/jamiec7919/vermeer/material"
	m "github.com/jamiec7919/vermeer/math"
)

const CHECK_EMPTY_LEAF = true

const VisRayEpsilon float32 = 0.0001

type TraverseElem struct {
	Node int32
	T    float32
}

// Heap allocation guarantees 16-byte alignment, stack allocation doesn't!
// 512 bytes structure
type TraceSupport struct {
	T     [4]float32
	Hits  [4]int32
	Stack [60]TraverseElem // Traverse Elem is 8 bytes
}

// 2 cache lines currently, or would be if 64 byte aligned, as it is should be 32byte aligned by
// Go allocator so still not too bad.
// We store the error offset in the result as we may want to move the point to either side depending
// on the material (e.g. refraction will need point on other side)
type RayResult struct {
	P        m.Vec3
	POffset  m.Vec3    // Offset to make sure any intersection point is outside face
	Ng, T, B m.Vec3    // Tangent space, Tg & Bg not normalized, Ng is normalized (as stored with face)
	Ns       m.Vec3    // Not normalized
	MtlId    int32     // 64 bytes (first line)
	UV       [4]m.Vec2 // 8 floats (48 bytes)
	Pu, Pv   [4]m.Vec3 // 12 float32
	Extra    map[string]interface{}
}

// 64bytes (one cache line)
type Ray struct {
	P, D     m.Vec3  // 6
	Dinv     m.Vec3  // 16
	Tclosest float32 // offset 4*9 = 36 (total in RayData is 512+36=548)

	S          [3]float32 //10
	Kx, Ky, Kz int32      // 13
}

type RayStats struct {
	Nnodes int
}

// For some reason Supp.T is not being aligned to 16-bytes..
// Aligned into 64byte blocks
type RayData struct {
	Supp   TraceSupport
	Ray    Ray
	Result RayResult
	//TransformRay Ray       // No longer used, all geometry is transformed to world space
	LocalToWorld m.Matrix4 // No longer used, to transform from Ray space back to world
	Stats        RayStats
}

func (r *RayData) InitRay(P, D m.Vec3) {
	r.Ray.P = P
	r.Ray.D = D
	r.Ray.Tclosest = m.Inf(1)
	r.Ray.setup()
}

func (r *RayData) InitVisRay(P0, P1 m.Vec3) {
	r.Ray.P = P0
	r.Ray.D = m.Vec3Sub(P1, P0)
	r.Ray.Tclosest = 1 - VisRayEpsilon
	r.Ray.setup()
}

var ErrNoHit = errors.New("No hit")

// r is initialized as vis ray, returns true if P1 is visible from P0.
func (r *RayData) IsVis() bool {
	if r.Ray.Tclosest < 1-VisRayEpsilon {
		return false
	} else {
		return true
	}
}

func (r *RayData) GetHitSurface(surface *material.SurfacePoint) error {
	if r.Ray.Tclosest < m.Inf(1) {
		surface.P = r.Result.P
		surface.N = r.Result.Ng
		//		surface.B = r.Result.Bg
		//		surface.T = r.Result.Tg
		for k := range surface.UV {
			surface.UV[k] = r.Result.UV[k]
			surface.Pu[k] = r.Result.Pu[k]
			surface.Pv[k] = r.Result.Pv[k]
		}

		surface.SetupTangentSpace(r.Result.Ns)
		//		surface.Ns = m.Vec3Normalize(r.Result.Ns)
		//		surface.B = m.Vec3Normalize(m.Vec3Cross(surface.Ns, surface.Pu[0]))
		//		surface.T = m.Vec3Normalize(m.Vec3Cross(surface.Ns, surface.B))

		//log.Printf("%v %v", surface.T, surface.B)
		surface.POffset = r.Result.POffset
		surface.MtlId = r.Result.MtlId
		//surface.Extra = copy(r.Result.Extra)
		return nil
	}

	return ErrNoHit
}

func (r *Ray) setup() {

	r.Kz = 0

	if m.Abs(r.D[1]) > m.Abs(r.D[2]) {
		if m.Abs(r.D[1]) > m.Abs(r.D[0]) {
			r.Kz = 1

		}
	} else {
		if m.Abs(r.D[2]) > m.Abs(r.D[0]) {
			r.Kz = 2
		}

	}

	r.Kx = r.Kz + 1

	if r.Kx == 3 {
		r.Kx = 0
	}

	r.Ky = r.Kx + 1

	if r.Ky == 3 {
		r.Ky = 0
	}

	if r.D[r.Kz] < 0.0 {
		tmp := r.Kx
		r.Kx = r.Ky
		r.Ky = tmp
	}

	// Divisions as accurate as possible
	z := float64(r.D[r.Kz])

	r.S[2] = float32(1.0 / z)
	r.S[0] = float32(float64(r.D[r.Kx]) / z)
	r.S[1] = float32(float64(r.D[r.Ky]) / z)

	r.Dinv[0] = float32(1.0 / float64(r.D[0]))
	r.Dinv[1] = float32(1.0 / float64(r.D[1]))
	r.Dinv[2] = float32(1.0 / float64(r.D[2]))

}
