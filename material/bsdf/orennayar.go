// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsdf

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/math/sample"
	"math"
)

// OrenNayar2 implements the Oren-Nayar diffuse microfacet model.
// Instanced for each point
type OrenNayar2 struct {
	Lambda    float32
	OmegaI    m.Vec3
	Roughness float32
}

// NewOrenNayar returns a new instance of the model for the given parameters.
func NewOrenNayar(sg *core.ShaderGlobals, roughness float32) *OrenNayar2 {
	return &OrenNayar2{sg.Lambda, sg.ViewDirection(), roughness * roughness}
}

// Sample implements core.BSDF.
func (b *OrenNayar2) Sample(r0, r1 float64) m.Vec3 {
	return sample.CosineHemisphere(r0, r1)
}

// PDF implements core.BSDF.
func (b *OrenNayar2) PDF(omegaO m.Vec3) float64 {
	ODotN := float64(m.Abs(omegaO[2]))

	return ODotN / math.Pi
}

// Eval implements core.BSDF.
func (b *OrenNayar2) Eval(omegaO m.Vec3) (rho colour.Spectrum) {

	sigma := b.Roughness

	A := 1 - (0.5 * (sigma * sigma) / ((sigma * sigma) + 0.57))

	B := 0.45 * (sigma * sigma) / ((sigma * sigma) + 0.09)

	phiI := m.Atan2(b.OmegaI[1], b.OmegaI[0])
	phiO := m.Atan2(omegaO[1], omegaO[0])
	thetaI := m.Acos(b.OmegaI[2])
	thetaO := m.Acos(omegaO[2])

	alpha := m.Max(thetaI, thetaO)
	beta := m.Min(thetaI, thetaO)

	C := m.Sin(alpha) * m.Tan(beta)

	gamma := m.Cos(phiO - phiI)
	// This might be ok:
	// gamma := dot(eyeDir - normal * dot(eyeDir, normal), light.direction - normal * dot(light.direction, normal))

	//*out = b.Kd
	scale := omegaO[2] * (A + (B * m.Max(0, gamma) * C))

	rho.Lambda = b.Lambda
	rho.FromRGB(colour.RGB{1, 1, 1})
	rho.Scale(scale / math.Pi)

	return
}
