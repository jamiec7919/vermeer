// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

// Ray type bit flags.
const (
	RayCamera = (1 << iota)
	RayShadow
)

// CheckEmptyLeaf is a debug constant. If set then empty leafs are explicitly checked.
const CheckEmptyLeaf = true

// VisRayEpsilon is the epsilon to use for shadow rays.
const VisRayEpsilon float32 = 0.0001

// TraverseElem is the stack element for traversal.
type TraverseElem struct {
	Node int32
	T    float32
}

// TraceSupport has the stacks and aligned data for SSE routines.
// Heap allocation guarantees 16-byte alignment, stack allocation doesn't!
// 512 + (32*8) bytes structure
type TraceSupport struct {
	T             [4]float32
	Hits          [4]int32
	Boxes         [4 * 3 * 2]float32
	Stack         [60]TraverseElem // Traverse Elem is 8 bytes
	TopLevelStack [32]TraverseElem
}

// RayResult holds the data from a ray intersection.
//
// 2 cache lines currently, or would be if 64 byte aligned, as it is should be 32byte aligned by
// Go allocator so still not too bad.
// We store the error offset in the result as we may want to move the point to either side depending
// on the material (e.g. refraction will need point on other side).
type RayResult struct {
	P        m.Vec3  // 12
	POffset  m.Vec3  // 12 Offset to make sure any intersection point is outside face
	Ng, T, B m.Vec3  // 36 Tangent space, Tg & Bg not normalized, Ng is normalized (as stored with face)
	Ns       m.Vec3  // 12 Not normalized
	MtlID    int32   // 4   64 bytes (first line)
	UV       m.Vec2  // 8 floats (48 bytes)
	Pu, Pv   m.Vec3  // 12 float32
	Bu, Bv   float32 // Barycentric coords
	Prim     Primitive
	ElemID   uint32
}

// Ray represents a ray in world space plus precalculated intersection values.
//
// 64bytes (one cache line).
type Ray struct {
	P, D     m.Vec3  // 6
	Dinv     m.Vec3  // 16
	Tclosest float32 // offset 4*9 = 36 (total in RayData is 512+36=548)

	S          [3]float32 //10
	Kx, Ky, Kz int32      // 13
}

// RayStats collects stats about the ray traversal.
//
// Deprecated: RayStats is unused.
type RayStats struct {
	Nnodes int
}

// RayData represents a ray plus support and transform data.  Will be refactored.
// For some reason Supp.T is not being aligned to 16-bytes.. (needs to be heap allocated)
// Aligned into 64byte blocks
type RayData struct {
	Supp         TraceSupport
	Ray          Ray
	SavedRay     Ray       // Saved version of ray if needed (i.e. we've transformed Ray)
	LocalToWorld m.Matrix4 // Local to world transform
	Result       RayResult
	Stats        RayStats
	Level        uint8 // recursion level
	rnd          *rand.Rand
	Lambda       float32
	Time         float32
	Type         uint32
}

// Init sets up the ray.  ty should be bitwise combination of RAY_ constants.  P is the
// start point and D is the direction.  maxdist is the length of the ray.  sg is used
// to get the Lambda, rng and Time parameters.
func (r *RayData) Init(ty uint32, P, D m.Vec3, maxdist float32, sg *ShaderGlobals) {
	r.Ray.P = P
	r.Ray.D = D
	r.Ray.Tclosest = maxdist
	r.Type = ty
	r.Ray.setup()
	r.Level = sg.Depth
	r.rnd = sg.rnd
	r.Lambda = sg.Lambda
	r.Time = sg.Time

}

// IsVis returns true if P1 is visible from P0.
// r is initialized as vis ray
func (r *RayData) IsVis() bool {
	if r.Ray.Tclosest < 1-VisRayEpsilon {
		return false
	}

	return true

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
