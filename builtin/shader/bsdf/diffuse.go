// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package bsdf provides built-in B(R/S)DF (Bidirectional Reflectance/Scattering Distribution Function)
models for Vermeer. */
package bsdf

import (
	"github.com/jamiec7919/vermeer/colour"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/math/sample"
	"math"
)

// Weight = f * dot(omegaO,n) / pdf
// pdf := dot(omegaO,n) / m.Pi
// f := rho / m.Pi

// Instanced for each point

// Lambert is basic lambertian diffuse.
type Lambert struct {
	Lambda  float32
	OmegaI  m.Vec3
	U, V, N m.Vec3
}

// NewLambert will return an instance of the model for the given point.
func NewLambert(lambda float32, omegaI m.Vec3, U, V, N m.Vec3) *Lambert {
	return &Lambert{lambda, m.Vec3BasisProject(U, V, N, omegaI), U, V, N}
}

// Sample implements core.BSDF.
func (b *Lambert) Sample(r0, r1 float64) m.Vec3 {
	return m.Vec3BasisExpand(b.U, b.V, b.N, sample.CosineHemisphere(r0, r1))
}

// PDF implements core.BSDF.
func (b *Lambert) PDF(_omegaO m.Vec3) float64 {
	omegaO := m.Vec3BasisProject(b.U, b.V, b.N, _omegaO)
	ODotN := float64(m.Max(omegaO[2], 0))

	if ODotN > 1.0 {
		ODotN = 1
	}

	return ODotN / math.Pi

}

// Eval implements core.BSDF.
func (b *Lambert) Eval(_omegaO m.Vec3) (rho colour.Spectrum) {
	omegaO := m.Vec3BasisProject(b.U, b.V, b.N, _omegaO)
	weight := m.Max(0, omegaO[2])

	rho.Lambda = b.Lambda
	rho.FromRGB(colour.RGB{1, 1, 1})
	rho.Scale(weight / m.Pi)

	return
}
