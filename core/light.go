// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

// Light represents a light that can be sampled by the system.
type Light interface {
	//	SamplePoint(*rand.Rand, *SurfacePoint, *float64) error                                // Sample a point on the surface

	// SampleArea samples a point on the surface of the light by area.
	// Returns nil on successful sample.
	SampleArea(*ShaderGlobals) error

	//	SampleDirection(*SurfacePoint, *rand.Rand, *m.Vec3, *colour.Spectrum, *float64) error // Sample direction given point

	// DiffuseShadeMult returns the diffuse lighting multiplier.
	DiffuseShadeMult() float32
}
