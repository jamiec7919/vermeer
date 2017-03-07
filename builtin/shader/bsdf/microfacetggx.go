// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsdf

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	//"log"
	"math"
)

// MicrofacetGGX implements the GGX specular microfacet model.
// Instanced for each point
type MicrofacetGGX struct {
	Lambda    float32
	OmegaR    m.Vec3 // reflected (view or out) direction
	Roughness float32
	Fresnel   core.Fresnel
	U, V, N   m.Vec3 // Tangent space

}

func chi(x float32) float32 {
	if x > 0.0 {
		return 1
	}
	return 0
}

func ggxSmithG1(omega, omegaM m.Vec3, alpha float32) float32 {

	// 2 / (1+sqrt(1+alpha^2*tan^2 theta_v))
	// tan^2(x) + 1 = sec^2(x)
	// sec(x) = 1/cos(x)
	// sec^2(x) = 1/cos_2(x)
	// tan^2(x) = 1/cos_2(x) - 1

	ODotN := m.Vec3Dot(omega, omegaM)

	//thetaV := m.Acos(omega[2])
	//tan := m.Tan(thetaV)
	//log.Printf("%v %v", tan*tan, (1.0/(omega[2]*omega[2]))-1)

	denom := 1 + m.Sqrt(1+(alpha*alpha)*((1.0/(omega[2]*omega[2]))-1))

	return chi(ODotN/omega[2]) * 2 / denom
}

func ggxD(omegaM m.Vec3, alpha float32) float32 {

	numer := alpha * alpha * chi(omegaM[2])

	if omegaM[2] == 1.0 {
		// if omegaM == {0,0,1} with alpha small there is a numerical problem
		// calculating the weight. Since this mostly happens with omegaM being chosen as the
		// perfect mirror direction (same as shade normal) we do the calculation directly here avoiding
		// the extra squaring of alpha.
		// denom = m.Pi * sqr32(alpha*alpha)
		return 1.0 / (m.Pi * alpha * alpha)
	}

	denom := m.Pi * sqr32(omegaM[2]*omegaM[2]) * sqr32(alpha*alpha+((1.0/(omegaM[2]*omegaM[2]))-1))

	//log.Printf("%v %v %v", omegaM, numer, denom)
	return numer / denom
}

func sign(v float32) float32 {
	if v < 0 {
		return -1
	}

	return 1

}

// NewMicrofacetGGX returns a new instance of the model for the given parameters.
func NewMicrofacetGGX(sg *core.ShaderContext, omegaI m.Vec3, fresnel core.Fresnel, roughness float32, U, V, N m.Vec3) *MicrofacetGGX {
	return &MicrofacetGGX{sg.Lambda, m.Vec3BasisProject(U, V, N, omegaI), roughness * roughness, fresnel, U, V, N}
}

// Sample implements core.BSDF.
func (b *MicrofacetGGX) Sample(r0, r1 float64) (omegaO m.Vec3) {

	alpha := sqr32(b.Roughness)

	thetaM := math.Atan2(float64(alpha)*math.Sqrt(r0), math.Sqrt(1-r0))
	phiM := 2.0 * math.Pi * r1

	omegaM := m.Vec3{m.Sin(float32(thetaM)) * m.Cos(float32(phiM)),
		m.Sin(float32(thetaM)) * m.Sin(float32(phiM)),
		m.Cos(float32(thetaM))}

	omegaO = m.Vec3Sub(m.Vec3Scale(2.0*m.Vec3DotAbs(omegaM, b.OmegaR), omegaM), b.OmegaR)

	return m.Vec3BasisExpand(b.U, b.V, b.N, m.Vec3Normalize(omegaO))
}

// PDF implements core.BSDF.
func (b *MicrofacetGGX) PDF(_omegaO m.Vec3) float64 {
	omegaO := m.Vec3BasisProject(b.U, b.V, b.N, _omegaO)

	alpha := sqr32(b.Roughness)

	var omegaM m.Vec3

	omegaM = m.Vec3Scale(sign(b.OmegaR[2]), m.Vec3Normalize(m.Vec3Add(b.OmegaR, omegaO)))

	//log.Printf("D: %v", ggxD(omegaM, alpha))
	return float64(ggxD(omegaM, alpha) * omegaM[2])
}

// Eval implements core.BSDF.
func (b *MicrofacetGGX) Eval(_omegaO m.Vec3) (rho colour.Spectrum) {
	omegaI := m.Vec3BasisProject(b.U, b.V, b.N, _omegaO)

	alpha := sqr32(b.Roughness)

	h := m.Vec3Scale(sign(b.OmegaR[2]), m.Vec3Normalize(m.Vec3Add(b.OmegaR, omegaI)))

	fresnel := b.Fresnel.Kr(m.Vec3DotAbs(b.OmegaR, h))

	numer := ggxSmithG1(b.OmegaR, h, alpha) * ggxSmithG1(omegaI, h, alpha) * ggxD(h, alpha)
	denom := 4 * m.Abs(b.OmegaR[2]) * m.Abs(omegaI[2])

	rho.Lambda = b.Lambda
	rho.FromRGB(fresnel)
	rho.Scale(m.Abs(omegaI[2]) * numer / denom)
	return
}

// This computes the weight as per the paper, but not sure it's useful for Vermeer.
func (b *MicrofacetGGX) _weight(omegaI m.Vec3) (rho colour.Spectrum) {
	alpha := sqr32(b.Roughness)

	var omegaM m.Vec3

	//etaI := float32(1)
	//etaO := b.IOR

	//if b.OmegaR[2] < 0.0 {
	//	etaI, etaO = etaO, etaI
	//}

	weight := float32(0)

	//if b.OmegaR[2] > 0 && omegaI[2] < 0 {
	//	omegaM = m.Vec3Normalize(m.Vec3Add(m.Vec3Scale(1.0, b.OmegaR), m.Vec3Scale(b.IOR, omegaI)))
	//} else if b.OmegaR[2] < 0 && omegaI[2] > 0 {
	//	omegaM = m.Vec3Normalize(m.Vec3Add(m.Vec3Scale(b.IOR, b.OmegaR), m.Vec3Scale(1.0, omegaI)))
	//} else {
	omegaM = m.Vec3Scale(sign(m.Vec3Dot(b.OmegaR, omegaI)), m.Vec3Normalize(m.Vec3Add(b.OmegaR, omegaI)))
	//	}

	weight = m.Vec3DotAbs(omegaI, omegaM) * ggxSmithG1(omegaI, omegaM, alpha) * ggxSmithG1(b.OmegaR, omegaM, alpha)
	weight /= m.Abs(omegaM[2]) * m.Abs(omegaI[2])

	rho.Lambda = b.Lambda
	rho.FromRGB(colour.RGB{1, 1, 1})
	rho.Scale(weight)
	return
}
