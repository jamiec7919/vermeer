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
	Lambda float32
	OmegaR m.Vec3 // reflected (view or out) direction
	//IOR          float32 // n_i/n_t  (n_i = air)
	Roughness    float32
	Fresnel      core.Fresnel
	transmissive bool
	thin         bool
	metal        bool
}

func chi(x float32) float32 {
	if x > 0.0 {
		return 1
	}
	return 0
}

func ggx_SmithG1(omega, omegaM m.Vec3, alpha float32) float32 {

	// 2 / (1+sqrt(1+alpha^2*tan^2 theta_v))
	// tan^2(x) + 1 = sec^2(x)
	// sec(x) = 1/cos(x)
	// sec^2(x) = 1/cos_2(x)
	// tan^2(x) = 1/cos_2(x) - 1

	o_dot_n := m.Vec3Dot(omega, omegaM)

	//thetaV := m.Acos(omega[2])
	//tan := m.Tan(thetaV)
	//log.Printf("%v %v", tan*tan, (1.0/(omega[2]*omega[2]))-1)

	denom := 1 + m.Sqrt(1+(alpha*alpha)*((1.0/(omega[2]*omega[2]))-1))

	return chi(o_dot_n/omega[2]) * 2 / denom
}

func ggx_D(omega_m m.Vec3, alpha float32) float32 {

	numer := alpha * alpha * chi(omega_m[2])

	var denom float32
	if omega_m[2] == 1.0 {
		// if omegaM == {0,0,1} with alpha small there is a numerical problem
		// calculating the weight. Since this mostly happens with omegaM being chosen as the
		// perfect mirror direction (same as shade normal) we do the calculation directly here avoiding
		// the extra squaring of alpha.
		// denom = m.Pi * sqr32(alpha*alpha)
		return 1.0 / (m.Pi * alpha * alpha)
	} else {
		denom = m.Pi * sqr32(omega_m[2]*omega_m[2]) * sqr32(alpha*alpha+((1.0/(omega_m[2]*omega_m[2]))-1))
	}
	//log.Printf("%v %v %v", omega_m, numer, denom)
	return numer / denom
}

func sign(v float32) float32 {
	if v < 0 {
		return -1
	} else {
		return 1
	}
}

// NewMicrofacetGGX returns a new instance of the model for the given parameters.
func NewMicrofacetGGX(sg *core.ShaderGlobals, fresnel core.Fresnel, roughness float32, transmissive, thin bool) *MicrofacetGGX {
	return &MicrofacetGGX{sg.Lambda, sg.ViewDirection(), roughness * roughness, fresnel, transmissive, thin, false}
}

// Sample implements core.BSDF.
func (b *MicrofacetGGX) Sample(r0, r1 float64) (omegaO m.Vec3) {

	alpha := sqr32(b.Roughness)

	theta_m := math.Atan2(float64(alpha)*math.Sqrt(r0), math.Sqrt(1-r0))
	phi_m := 2.0 * math.Pi * r1

	omega_m := m.Vec3{m.Sin(float32(theta_m)) * m.Cos(float32(phi_m)),
		m.Sin(float32(theta_m)) * m.Sin(float32(phi_m)),
		m.Cos(float32(theta_m))}

	omegaO = m.Vec3Sub(m.Vec3Scale(2.0*m.Vec3DotAbs(omega_m, b.OmegaR), omega_m), b.OmegaR)

	return m.Vec3Normalize(omegaO)
}

// PDF implements core.BSDF.
func (b *MicrofacetGGX) PDF(omega_i m.Vec3) float64 {
	alpha := sqr32(b.Roughness)

	var omegaM m.Vec3

	omegaM = m.Vec3Scale(sign(b.OmegaR[2]), m.Vec3Normalize(m.Vec3Add(b.OmegaR, omega_i)))

	//log.Printf("D: %v", ggx_D(omegaM, alpha))
	return float64(ggx_D(omegaM, alpha) * omegaM[2])
}

// Eval implements core.BSDF.
func (b *MicrofacetGGX) Eval(omega_i m.Vec3) (rho colour.Spectrum) {
	alpha := sqr32(b.Roughness)

	h := m.Vec3Scale(sign(b.OmegaR[2]), m.Vec3Normalize(m.Vec3Add(b.OmegaR, omega_i)))

	fresnel := b.Fresnel.Kr(m.Vec3DotAbs(b.OmegaR, h))

	numer := ggx_SmithG1(b.OmegaR, h, alpha) * ggx_SmithG1(omega_i, h, alpha) * ggx_D(h, alpha)
	denom := 4 * m.Abs(b.OmegaR[2]) * m.Abs(omega_i[2])

	rho.Lambda = b.Lambda
	rho.FromRGB(fresnel[0], fresnel[1], fresnel[2])
	rho.Scale(m.Abs(omega_i[2]) * numer / denom)
	return
}

// This computes the weight as per the paper, but not sure it's useful for Vermeer.
func (b *MicrofacetGGX) _weight(omega_i m.Vec3) (rho colour.Spectrum) {
	alpha := sqr32(b.Roughness)

	var omegaM m.Vec3

	//etaI := float32(1)
	//etaO := b.IOR

	//if b.OmegaR[2] < 0.0 {
	//	etaI, etaO = etaO, etaI
	//}

	weight := float32(0)

	//if b.OmegaR[2] > 0 && omega_i[2] < 0 {
	//	omegaM = m.Vec3Normalize(m.Vec3Add(m.Vec3Scale(1.0, b.OmegaR), m.Vec3Scale(b.IOR, omega_i)))
	//} else if b.OmegaR[2] < 0 && omega_i[2] > 0 {
	//	omegaM = m.Vec3Normalize(m.Vec3Add(m.Vec3Scale(b.IOR, b.OmegaR), m.Vec3Scale(1.0, omega_i)))
	//} else {
	omegaM = m.Vec3Scale(sign(m.Vec3Dot(b.OmegaR, omega_i)), m.Vec3Normalize(m.Vec3Add(b.OmegaR, omega_i)))
	//	}

	weight = m.Vec3DotAbs(omega_i, omegaM) * ggx_SmithG1(omega_i, omegaM, alpha) * ggx_SmithG1(b.OmegaR, omegaM, alpha)
	weight /= m.Abs(omegaM[2]) * m.Abs(omega_i[2])

	rho.Lambda = b.Lambda
	rho.FromRGB(1, 1, 1)
	rho.Scale(weight)
	return
}
