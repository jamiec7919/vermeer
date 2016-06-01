// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsdf

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
)

// SpecularTransmission implements perfect specular transmission.
// Instanced for each point
type SpecularTransmission struct {
	Lambda  float32
	OmegaR  m.Vec3
	ior     float32
	fresnel core.Fresnel
	thin    bool
}

// NewSpecularTransmission returns a new instance of the model.
func NewSpecularTransmission(sg *core.ShaderGlobals, ior float32, fresnel core.Fresnel, thin bool) *SpecularTransmission {
	return &SpecularTransmission{sg.Lambda, sg.ViewDirection(), ior, fresnel, thin}
}

// Sample implements core.BSDF.
func (b *SpecularTransmission) Sample(r0, r1 float64) (omegaO m.Vec3) {

	//fresnel := fresnel.Kr(m.Vec3DotAbs(b.OmegaRm.Vec3{0, 0, 1})).Maxh()

	if b.thin {
		omegaO = m.Vec3Neg(b.OmegaR)
	} else {
		omegaO = refract(b.OmegaR, b.ior)
	}

	omegaO = m.Vec3Normalize(omegaO)
	return
}

// PDF implements core.BSDF.
func (b *SpecularTransmission) PDF(omega_o m.Vec3) float64 {
	return 1
}

// Eval implements core.BSDF.
func (b *SpecularTransmission) Eval(omegaO m.Vec3) (rho colour.Spectrum) {
	//	fresnel := DielectricFresnel(b.OmegaR, m.Vec3{0, 0, 1}, b.ior)

	//	pdf := fresnel
	//	if sign(omegaO[2]) != sign(b.OmegaR[2]) {
	//		// refracted
	//		pdf = 1 - pdf
	//	}

	fresnel := b.fresnel.Kr(m.Vec3DotAbs(b.OmegaR, m.Vec3{0, 0, 1}))

	rho.Lambda = b.Lambda
	rho.FromRGB(1-fresnel[0], 1-fresnel[1], 1-fresnel[2])
	rho.Scale(m.Vec3DotAbs(omegaO, m.Vec3{0, 0, 1}))
	return
}
