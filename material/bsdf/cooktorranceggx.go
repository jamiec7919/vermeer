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

// CookTorranceGGX2 is a microfacet model..
//
// Deprecated: pending a rewrite that actually works.  Use
// MicrofacetGGX for now.
// Instanced for each point
type CookTorranceGGX2 struct {
	Lambda    float32
	OmegaI    m.Vec3
	IOR       float32 // n_i/n_t  (n_i = air)
	Roughness float32
}

// NewCookTorranceGGX returns a new instance of the model.
//
// Deprecated:
func NewCookTorranceGGX(sg *core.ShaderGlobals, IOR, roughness float32) *CookTorranceGGX2 {
	return &CookTorranceGGX2{sg.Lambda, sg.ViewDirection(), IOR, roughness}
}

// Sample returns a sample.
//
// Deprecated:
func (b *CookTorranceGGX2) Sample(r0, r1 float64) m.Vec3 {
	alpha := sqr32(b.Roughness)

	if alpha == 0.0 {
		alpha = 0.001
	}
	omegaI := b.OmegaI
	omegaM := b.sample(m.Vec3Normalize(omegaI), float64(alpha), float64(alpha), r0, r1)

	omegaO := m.Vec3Normalize(m.Vec3Sub(m.Vec3Scale(2.0*m.Vec3DotAbs(omegaI, omegaM), omegaM), omegaI))

	//	log.Printf("omegaO: %v", omegaO)
	return omegaO
}

// PDF returns the pdf for the given sample.
//
// Deprecated:
func (b *CookTorranceGGX2) PDF(omegaO m.Vec3) float64 {
	alpha := sqr32(b.Roughness)

	if alpha == 0.0 {
		alpha = 0.001
	}

	D := func(h m.Vec3, alpha float64) float64 {
		HDotN := float64(h[2])
		return sqr(alpha) / (math.Pi * sqr((HDotN*HDotN)*(sqr(alpha)-1.0)+1.0))
	}
	omegaI := b.OmegaI
	//IDotN := float64(omegaI[2])
	ODotN := float64(omegaO[2])

	omegaM := m.Vec3Normalize(m.Vec3Add(omegaI, omegaO))
	pdf := float64(b.G1(omegaO, alpha)*m.Vec3DotAbs(omegaO, omegaM)) * D(omegaM, float64(alpha)) / math.Abs(ODotN)

	//	log.Printf("%v", pdf)
	return pdf
}

// Eval returns the.
//
// Deprecated:
func (b *CookTorranceGGX2) Eval(omegaO m.Vec3) (rho colour.Spectrum) {
	alpha := sqr32(b.Roughness)

	if alpha == 0.0 {
		alpha = 0.001
	}

	omegaI := b.OmegaI
	D := func(h m.Vec3, alpha float32) float32 {
		HDotN := h[2]

		return sqr32(alpha) / (m.Pi * sqr32(float32(float64(HDotN*HDotN)*(float64(alpha)*float64(alpha)-1)+1)))
	}

	F := func(i, h m.Vec3) float32 {
		c := m.Abs(m.Vec3Dot(i, h))
		IOR := b.IOR
		g := m.Sqrt(sqr32((IOR) - 1 + c*c))

		return 0.5 * (sqr32(g-c) / sqr32(g+c)) * (1 + (sqr32(c*(g+c)-1) / sqr32(c*(g-c)+1)))
	}

	IDotN := omegaI[2]
	ODotN := omegaO[2]

	omegaM := m.Vec3Normalize(m.Vec3Add(omegaI, omegaO))

	spec := D(omegaM, alpha) * F(omegaO, omegaM) * b.G1(omegaI, alpha) * b.G1(omegaO, alpha) / (4.0 * IDotN * ODotN)

	rho.Lambda = b.Lambda
	rho.FromRGB(colour.RGB{1, 1, 1})
	rho.Scale(spec)
	return
}

func sqr(x float64) float64   { return x * x }
func sqr32(x float32) float32 { return x * x }

// G1 is the Smith shadowing function.
func (b *CookTorranceGGX2) G1(omega m.Vec3, alpha float32) float32 {

	// 2 / (1+sqrt(1+alpha^2*tan^2 theta_v))
	// tan^2(x) + 1 = sec^2(x)
	// sec(x) = 1/cos(x)
	// sec^2(x) = 1/cos_2(x)
	// tan^2(x) = 1/cos_2(x) - 1

	ODotN := omega[2]
	denom := 1 + m.Sqrt(1+(alpha*alpha)*((1.0/(ODotN*ODotN))-1))

	if denom == 0.0 {
		log.Printf("denom %v %v", omega, alpha)
	}
	return 2 / denom
}

func (b *CookTorranceGGX2) sample(omegaI m.Vec3, alphaX, alphaY, r0, r1 float64) (omegaM m.Vec3) {
	omegaI64 := [3]float64{
		alphaX * float64(omegaI[0]),
		alphaY * float64(omegaI[1]),
		float64(omegaI[2]),
	}

	invOmegaI := 1.0 / math.Sqrt(omegaI64[0]*omegaI64[0]+omegaI64[1]*omegaI64[1]+omegaI64[2]*omegaI64[2])
	omegaI64[0] = invOmegaI * omegaI64[0]
	omegaI64[1] = invOmegaI * omegaI64[1]
	omegaI64[2] = invOmegaI * omegaI64[2]

	theta := float64(0)
	phi := float64(0)

	if omegaI64[2] < 0.99999 {
		theta = math.Acos(omegaI64[2])
		phi = math.Atan2(omegaI64[1], omegaI64[0])
	}

	if math.IsNaN(theta) {
		log.Printf("theta %v", theta)

	}
	if math.IsNaN(phi) {
		log.Printf("phi %v", phi)
	}
	slopeX, slopeY := b.sample11(theta, r0, r1)
	//	slopeX, slopeY := float32(slopeX_), float32(slopeY_)
	if math.IsNaN(slopeX) {
		log.Printf("slopeX %v %v %v", slopeX, theta, phi)
	}
	if math.IsNaN(slopeY) {
		log.Printf("slopeY %v %v %v", slopeY, theta, phi)
	}

	tmp := math.Cos(phi)*slopeX - math.Sin(phi)*slopeY
	slopeY = math.Sin(phi)*slopeX + math.Cos(phi)*slopeY
	slopeX = tmp

	slopeX = alphaX * slopeX
	slopeY = alphaY * slopeY

	invOmegaM := math.Sqrt(slopeX*slopeX + slopeY*slopeY + 1.0)
	omegaM[0] = -float32(slopeX / invOmegaM)
	omegaM[1] = -float32(slopeY / invOmegaM)
	omegaM[2] = float32(1.0 / invOmegaM)
	return
}

func (b *CookTorranceGGX2) sample11(thetaI, r0, r1 float64) (slopeX, slopeY float64) {
	if thetaI < 0.0001 {
		r := math.Sqrt(r0 / (1 - r0))
		phi := 6.28318530718 * r1
		slopeX = r * math.Cos(phi)
		slopeY = r * math.Sin(phi)
		return
	}
	// precomputations
	tanThetaI := math.Tan(thetaI)
	a := 1.0 / (tanThetaI)
	G1 := 2.0 / (1 + math.Sqrt(1.0+1.0/(a*a)))

	// sample slopeX
	A := 2.0*r0/G1 - 1.0

	if A == 1.0 {
		// Randomly seem to be getting several times where r0 == G1 which
		// fails miserably below.  Maybe using doubles would be better after all!
		A = 0.999999
	} else if A == -1.0 {
		// Randomly sometimes r0 == 0.0
		A = -0.999999
	}

	tmp := 1.0 / (A*A - 1.0)
	B := tanThetaI
	D := math.Sqrt(B*B*tmp*tmp - (A*A-B*B)*tmp)

	slopeX1 := B*tmp - D
	slopeX2 := B*tmp + D

	if A < 0 || slopeX2 > 1.0/tanThetaI {
		slopeX = slopeX1
	} else {
		slopeX = slopeX2
	}

	if math.IsNaN(slopeX1) || math.IsNaN(slopeX2) {
		log.Printf("slope %v %v tan_theta:%v B:%v tmp:%v D:%v A:%v G1:%v a:%v r0:%v r1:%v", slopeX1, slopeX2, tanThetaI, B, tmp, D, A, G1, a, r0, r1)
	}

	S := float64(0)
	if r1 > 0.5 {
		S = 1.0
		r1 = 2.0 * (r1 - 0.5)
	} else {
		S = -1.0
		r1 = 2.0 * (0.5 - r1)
	}

	z := (r1 * (r1*(r1*0.27385-0.73369) + 0.46341)) / (r1*(r1*(r1*0.093073+0.309420)-1.000000) + 0.597999)

	slopeY = S * z * math.Sqrt(1.0+slopeX*slopeX)

	return
}
