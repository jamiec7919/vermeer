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
	RayTypeCamera uint32 = (1 << iota)
	RayTypeShadow
)

// ShadowRayEpsilon ensures no self-intersecting shadow rays.
const ShadowRayEpsilon float32 = 0.0001

// Ray is a mostly public structure used by individual intersection routines.
// These should only ever be created with ShaderContext.NewRay as they need to
// be in the pool.
type Ray struct {
	P, D       m.Vec3     // Position and direction
	Dinv       m.Vec3     // 1/direction
	Tclosest   float32    // t value of intersection (here instead of in Result to aid cache)
	S          [3]float32 // Precalculated ray-triangle members
	Kx, Ky, Kz int32      // Precalculated ray-triangle members

	// 64 bytes to here
	Time, Lambda float32 // Time value and wavelength
	X, Y         int32   // Raster position
	Sx, Sy       float32 // Screen space coords [-1,1]x[-1,1]
	Level        uint8
	Type         uint32 // Ray type bits
	I            int    // pixel sample index

	// Ray differentials
	DdPdx, DdPdy m.Vec3 // Ray differential
	DdDdx, DdDdy m.Vec3 // Ray differential

	NodesT, LeafsT int

	next *Ray // Pool list
	Task *RenderTask
}

// Init sets up the ray.  ty should be bitwise combination of RAY_ constants.  P is the
// start point and D is the direction.  maxdist is the length of the ray.  sg is used
// to get the Lambda, Time parameters and ray differntial calculations.
func (r *Ray) Init(ty uint32, P, D m.Vec3, maxdist float32, level uint8, lambda, time float32) {
	r.P = P
	r.D = D
	r.Tclosest = maxdist
	r.Type = ty
	r.setup()

	r.Level = level
	r.Lambda = lambda
	r.Time = time
	r.NodesT = 0
	r.LeafsT = 0
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

// RenderTask represents an image tile.
// One of these is allocated per goroutine so no worries about thread safety.
type RenderTask struct {
	Traversal struct {
		T        [4]float32         // Temporary T values, MUSE be aligned to 16 bytes
		Hits     [4]int32           // Temporary hit values, MUSE be aligned to 16 bytes
		Boxes    [4 * 3 * 2]float32 // Temporary boxes structure, MUST be aligned to 16 bytes
		StackTop int32              // This is the top of the stack for any traversal to start from
		Stack    [90]struct {
			T    float32
			Node int32
		}
	}
	rand    *rand.Rand
	rayPool *Ray
	cxtPool *ShaderContext
}

// NewRay allocates a ray from the pool.
func (rc *RenderTask) NewRay() *Ray {
	if rc.rayPool == nil {
		ray := new(Ray)
		ray.Task = rc
		return ray
	}

	ray := rc.rayPool
	rc.rayPool = rc.rayPool.next
	return ray

}

// ReleaseRay releases ray to pool.
func (rc *RenderTask) ReleaseRay(ray *Ray) {
	ray.next = rc.rayPool
	rc.rayPool = ray
}

// NewShaderContext allocates a ShaderContext from the pool.
func (rc *RenderTask) NewShaderContext() *ShaderContext {
	if rc.cxtPool == nil {
		sc := new(ShaderContext)
		sc.task = rc
		return sc
	}

	sc := rc.cxtPool
	rc.cxtPool = rc.cxtPool.next
	return sc

}

// ReleaseShaderContext releases context back to pool.
func (rc *RenderTask) ReleaseShaderContext(sc *ShaderContext) {
	sc.next = rc.cxtPool
	rc.cxtPool = sc
}
