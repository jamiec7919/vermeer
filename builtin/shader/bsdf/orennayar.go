// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsdf

import (
	"github.com/jamiec7919/vermeer/colour"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/math/sample"
	"math"
)

// OrenNayar implements the Oren-Nayar diffuse microfacet model.
// Instanced for each point
type OrenNayar struct {
	Lambda    float32
	OmegaI    m.Vec3
	Roughness float32
	U, V, N   m.Vec3 // Tangent space
}

// NewOrenNayar returns a new instance of the model for the given parameters.
func NewOrenNayar(lambda float32, omegaI m.Vec3, roughness float32, U, V, N m.Vec3) *OrenNayar {
	return &OrenNayar{lambda, m.Vec3BasisProject(U, V, N, omegaI), roughness * roughness, U, V, N}
}

// Sample implements core.BSDF.
func (b *OrenNayar) Sample(r0, r1 float64) m.Vec3 {
	return m.Vec3BasisExpand(b.U, b.V, b.N, sample.CosineHemisphere(r0, r1))
}

// PDF implements core.BSDF.
func (b *OrenNayar) PDF(_omegaO m.Vec3) float64 {
	omegaO := m.Vec3BasisProject(b.U, b.V, b.N, _omegaO)
	ODotN := float64(m.Max(0, omegaO.Z))

	return ODotN / math.Pi
}

// Eval implements core.BSDF.
func (b *OrenNayar) Eval(_omegaO m.Vec3) (rho colour.Spectrum) {

	omegaO := m.Vec3BasisProject(b.U, b.V, b.N, _omegaO)
	sigma := b.Roughness

	A := 1 - (0.5 * (sigma * sigma) / ((sigma * sigma) + 0.57))

	B := 0.45 * (sigma * sigma) / ((sigma * sigma) + 0.09)

	phiI := m.Atan2(b.OmegaI.Y, b.OmegaI.X)
	phiO := m.Atan2(omegaO.Y, omegaO.X)
	thetaI := m.Acos(b.OmegaI.Z)
	thetaO := m.Acos(omegaO.Z)

	alpha := m.Max(thetaI, thetaO)
	beta := m.Min(thetaI, thetaO)

	C := m.Sin(alpha) * m.Tan(beta)

	gamma := m.Cos(phiO - phiI)
	// This might be ok:
	// gamma := dot(eyeDir - normal * dot(eyeDir, normal), light.direction - normal * dot(light.direction, normal))

	//*out = b.Kd
	scale := omegaO.Z * (A + (B * m.Max(0, gamma) * C))

	rho.Lambda = b.Lambda
	rho.FromRGB(colour.RGB{1, 1, 1})
	rho.Scale(scale / math.Pi)

	return
}
