// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

type Light interface {
	//	SamplePoint(*rand.Rand, *SurfacePoint, *float64) error                                // Sample a point on the surface
	SampleArea(*ShaderGlobals) error // Sample a point on the surface visible from first point
	//	SampleDirection(*SurfacePoint, *rand.Rand, *m.Vec3, *colour.Spectrum, *float64) error // Sample direction given point

	DiffuseShadeMult() float32
}
