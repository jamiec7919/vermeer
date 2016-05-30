// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsdf

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"log"
	"math"
)

// MicrofacetGGX implements the GGX specular microfacet model.
// Instanced for each point
type MicrofacetGGX struct {
	Lambda    float32
	OmegaR    m.Vec3  // reflected (view or out) direction
	IOR       float32 // n_i/n_t  (n_i = air)
	Roughness float32
}

func ggx_SmithG1(omega m.Vec3, alpha float32) float32 {

	// 2 / (1+sqrt(1+alpha^2*tan^2 theta_v))
	// tan^2(x) + 1 = sec^2(x)
	// sec(x) = 1/cos(x)
	// sec^2(x) = 1/cos_2(x)
	// tan^2(x) = 1/cos_2(x) - 1

	o_dot_n := omega[2]
	denom := 1 + m.Sqrt(1+(alpha*alpha)*((1.0/(o_dot_n*o_dot_n))-1))

	if denom == 0.0 {
		log.Printf("denom %v %v", omega, alpha)
	}
	return 2 / denom
}

func ggx_D(omega_m m.Vec3, alpha float32) float32 {
	numer := alpha * alpha * 1.0

	denom := m.Pi * sqr32(omega_m[2]*omega_m[2]) * sqr32(alpha*alpha+((1.0/(omega_m[2]*omega_m[2]))-1))
	return numer / denom
}

// NewMicrofacetGGX returns a new instance of the model for the given parameters.
func NewMicrofacetGGX(sg *core.ShaderGlobals, IOR, roughness float32) *MicrofacetGGX {
	return &MicrofacetGGX{sg.Lambda, sg.ViewDirection(), IOR, roughness * roughness}
}

// Sample implements core.BSDF.
func (b *MicrofacetGGX) Sample(r0, r1 float64) m.Vec3 {
	alpha := sqr32(b.Roughness)

	theta_m := math.Atan2(float64(alpha)*math.Sqrt(r0), math.Sqrt(1-r0))
	phi_m := 2.0 * math.Pi * r1

	omega_m := m.Vec3{m.Sin(float32(theta_m)) * m.Cos(float32(phi_m)),
		m.Sin(float32(theta_m)) * m.Sin(float32(phi_m)),
		m.Cos(float32(theta_m))}

	if omega_m[2] < 0.0 {
		omega_m = m.Vec3Neg(omega_m)
	}
	omega_i := m.Vec3Sub(m.Vec3Scale(2.0*m.Vec3Dot(omega_m, b.OmegaR), omega_m), b.OmegaR)

	//log.Printf("%v %v", m.Vec3Length(omega_i), omega_i)

	return m.Vec3Normalize(omega_i)
}

// PDF implements core.BSDF.
func (b *MicrofacetGGX) PDF(omega_i m.Vec3) float64 {
	alpha := sqr32(b.Roughness)

	m := m.Vec3Normalize(m.Vec3Add(b.OmegaR, omega_i))

	return float64(ggx_D(m, alpha) * m[2])
}

// Eval implements core.BSDF.
func (b *MicrofacetGGX) Eval(omega_i m.Vec3) (rho colour.Spectrum) {
	alpha := sqr32(b.Roughness)
	omega_m := m.Vec3Normalize(m.Vec3Add(b.OmegaR, omega_i))

	weight := m.Vec3Dot(omega_i, omega_m) * ggx_SmithG1(omega_i, alpha) * ggx_SmithG1(b.OmegaR, alpha)
	weight /= omega_m[2] * omega_i[2]

	rho.Lambda = b.Lambda
	rho.FromRGB(1, 1, 1)
	rho.Scale(weight)
	return
}
