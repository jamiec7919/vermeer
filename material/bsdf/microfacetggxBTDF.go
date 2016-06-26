// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsdf

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"math"
)

// MicrofacetTransmissionGGX implements the GGX specular microfacet model.
// Instanced for each point
type MicrofacetTransmissionGGX struct {
	Lambda    float32
	OmegaR    m.Vec3 // reflected (view or out) direction
	Roughness float32
	fresnel   core.Fresnel
	ior       float32
	thin      bool
}

// NewMicrofacetTransmissionGGX returns a new instance of the model for the given parameters.
func NewMicrofacetTransmissionGGX(sg *core.ShaderGlobals, ior, roughness float32, fresnel core.Fresnel, thin bool) *MicrofacetTransmissionGGX {
	return &MicrofacetTransmissionGGX{sg.Lambda, sg.ViewDirection(), roughness * roughness, fresnel, ior, thin}
}

// Sample implements core.BSDF.
func (b *MicrofacetTransmissionGGX) Sample(r0, r1 float64) (omegaO m.Vec3) {

	alpha := sqr32(b.Roughness)

	thetaM := math.Atan2(float64(alpha)*math.Sqrt(r0), math.Sqrt(1-r0))
	phiM := 2.0 * math.Pi * r1

	omegaM := m.Vec3{m.Sin(float32(thetaM)) * m.Cos(float32(phiM)),
		m.Sin(float32(thetaM)) * m.Sin(float32(phiM)),
		m.Cos(float32(thetaM))}

	//		log.Printf("reflect")
	if b.thin {
		omegaO = m.Vec3Neg(b.OmegaR)
		return
	}

	//		log.Printf("transmit")
	eta := 1.0 / b.ior
	sign := sign(b.OmegaR[2])

	if sign < 0 {
		// Approaching from 'in media' so swap etaI & etaO
		eta = b.ior
	}

	c := m.Vec3Dot(b.OmegaR, omegaM)

	a := eta*c - sign*m.Sqrt(1.0+eta*(c*c-1.0))

	omegaO = m.Vec3Sub(m.Vec3Scale(a, omegaM), m.Vec3Scale(eta, b.OmegaR))

	return m.Vec3Normalize(omegaO)
}

// PDF implements core.BSDF.
func (b *MicrofacetTransmissionGGX) PDF(omegaO m.Vec3) float64 {

	if b.thin {
		return 1
	}

	alpha := sqr32(b.Roughness)

	etaI := float32(1)
	etaO := b.ior

	if b.OmegaR[2] < 0.0 {
		etaI, etaO = etaO, etaI
	}

	h := m.Vec3Normalize(m.Vec3Neg(m.Vec3Add(m.Vec3Scale(etaI, b.OmegaR), m.Vec3Scale(etaO, omegaO))))

	return float64(ggxD(h, alpha) * h[2])
}

// Eval implements core.BSDF.
func (b *MicrofacetTransmissionGGX) Eval(omegaO m.Vec3) (rho colour.Spectrum) {
	alpha := sqr32(b.Roughness)

	etaI := float32(1)
	etaO := b.ior

	if b.OmegaR[2] < 0.0 {
		etaI, etaO = etaO, etaI
	}

	h := m.Vec3Normalize(m.Vec3Neg(m.Vec3Add(m.Vec3Scale(etaI, b.OmegaR), m.Vec3Scale(etaO, omegaO))))

	factor1 := (m.Vec3DotAbs(b.OmegaR, h) * m.Vec3DotAbs(omegaO, h)) / (b.OmegaR[2] * omegaO[2])

	fresnel := b.fresnel.Kr(m.Vec3DotAbs(b.OmegaR, h))

	numer := sqr32(etaO) * ggxSmithG1(b.OmegaR, h, alpha) * ggxSmithG1(omegaO, h, alpha) * ggxD(h, alpha)

	denom := sqr32(etaI*m.Vec3Dot(b.OmegaR, h) + etaO*m.Vec3Dot(omegaO, h))

	weight := factor1 * (numer / denom)

	rho.Lambda = b.Lambda
	rho.FromRGB(1-fresnel[0], 1-fresnel[1], 1-fresnel[2])
	rho.Scale(m.Abs(omegaO[2]) * weight)
	return
}
