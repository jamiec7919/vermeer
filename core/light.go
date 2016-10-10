// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/jamiec7919/vermeer/colour"
	m "github.com/jamiec7919/vermeer/math"
)

// LightSample is used for direct lighting samples.
type LightSample struct {
	P      m.Vec3
	Weight float32
	Liu    colour.Spectrum
	Ld     m.Vec3
	Ldist  float32
}

// Light represents a light that can be sampled by the system.
type Light interface {

	// SampleArea samples n points on the surface of the light by area as seen from given ShaderGlobals point.
	// sg.Lsamples should be filled with the samples using append.
	// Returns nil on successful sample.
	SampleArea(sg *ShaderContext, n int) error

	// DiffuseShadeMult returns the diffuse lighting multiplier.
	DiffuseShadeMult() float32

	// NumSamples returns the number of samples that should be taken from this light for the
	// given context.
	NumSamples(sg *ShaderContext) int

	// Geom returns the Geom associated with this light.
	Geom() Geom
}
