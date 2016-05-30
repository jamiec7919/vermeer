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

//Deprecated: CookTorranceGGX2 is deprecated pending a rewrite that actually works.  Use
// MicrofacetGGX for now.
// Instanced for each point
type CookTorranceGGX2 struct {
	Lambda    float32
	OmegaI    m.Vec3
	IOR       float32 // n_i/n_t  (n_i = air)
	Roughness float32
}

//Deprecated:
func NewCookTorranceGGX(sg *core.ShaderGlobals, IOR, roughness float32) *CookTorranceGGX2 {
	return &CookTorranceGGX2{sg.Lambda, sg.ViewDirection(), IOR, roughness}
}

//Deprecated:
func (b *CookTorranceGGX2) Sample(r0, r1 float64) m.Vec3 {
	alpha := sqr32(b.Roughness)

	if alpha == 0.0 {
		alpha = 0.001
	}
	omega_i := b.OmegaI
	omega_m := b.sample(m.Vec3Normalize(omega_i), float64(alpha), float64(alpha), r0, r1)

	omega_o := m.Vec3Normalize(m.Vec3Sub(m.Vec3Scale(2.0*m.Vec3DotAbs(omega_i, omega_m), omega_m), omega_i))

	//	log.Printf("omega_o: %v", omega_o)
	return omega_o
}

//Deprecated:
func (b *CookTorranceGGX2) PDF(omega_o m.Vec3) float64 {
	alpha := sqr32(b.Roughness)

	if alpha == 0.0 {
		alpha = 0.001
	}

	D := func(h m.Vec3, alpha float64) float64 {
		h_dot_n := float64(h[2])
		return sqr(alpha) / (math.Pi * sqr((h_dot_n*h_dot_n)*(sqr(alpha)-1.0)+1.0))
	}
	omega_i := b.OmegaI
	//i_dot_n := float64(omega_i[2])
	o_dot_n := float64(omega_o[2])

	omega_m := m.Vec3Normalize(m.Vec3Add(omega_i, omega_o))
	pdf := float64(b.G1(omega_o, alpha)*m.Vec3DotAbs(omega_o, omega_m)) * D(omega_m, float64(alpha)) / math.Abs(o_dot_n)

	//	log.Printf("%v", pdf)
	return pdf
}

//Deprecated:
func (b *CookTorranceGGX2) Eval(omega_o m.Vec3) (rho colour.Spectrum) {
	alpha := sqr32(b.Roughness)

	if alpha == 0.0 {
		alpha = 0.001
	}

	omega_i := b.OmegaI
	D := func(h m.Vec3, alpha float32) float32 {
		h_dot_n := h[2]

		return sqr32(alpha) / (m.Pi * sqr32(float32(float64(h_dot_n*h_dot_n)*(float64(alpha)*float64(alpha)-1)+1)))
	}

	F := func(i, h m.Vec3) float32 {
		c := m.Abs(m.Vec3Dot(i, h))
		IOR := b.IOR
		g := m.Sqrt(sqr32((IOR) - 1 + c*c))

		return 0.5 * (sqr32(g-c) / sqr32(g+c)) * (1 + (sqr32(c*(g+c)-1) / sqr32(c*(g-c)+1)))
	}

	i_dot_n := omega_i[2]
	o_dot_n := omega_o[2]

	omega_m := m.Vec3Normalize(m.Vec3Add(omega_i, omega_o))

	spec := D(omega_m, alpha) * F(omega_o, omega_m) * b.G1(omega_i, alpha) * b.G1(omega_o, alpha) / (4.0 * i_dot_n * o_dot_n)

	rho.Lambda = b.Lambda
	rho.FromRGB(1, 1, 1)
	rho.Scale(spec)
	return
}

func sqr(x float64) float64   { return x * x }
func sqr32(x float32) float32 { return x * x }

func (b *CookTorranceGGX2) G1(omega m.Vec3, alpha float32) float32 {

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

func (b *CookTorranceGGX2) sample(omega_i m.Vec3, alpha_x, alpha_y, r0, r1 float64) (omega_m m.Vec3) {
	omega_i64 := [3]float64{
		alpha_x * float64(omega_i[0]),
		alpha_y * float64(omega_i[1]),
		float64(omega_i[2]),
	}

	inv_omega_i := 1.0 / math.Sqrt(omega_i64[0]*omega_i64[0]+omega_i64[1]*omega_i64[1]+omega_i64[2]*omega_i64[2])
	omega_i64[0] = inv_omega_i * omega_i64[0]
	omega_i64[1] = inv_omega_i * omega_i64[1]
	omega_i64[2] = inv_omega_i * omega_i64[2]

	theta := float64(0)
	phi := float64(0)

	if omega_i64[2] < 0.99999 {
		theta = math.Acos(omega_i64[2])
		phi = math.Atan2(omega_i64[1], omega_i64[0])
	}

	if math.IsNaN(theta) {
		log.Printf("theta %v", theta)

	}
	if math.IsNaN(phi) {
		log.Printf("phi %v", phi)
	}
	slope_x, slope_y := b.sample11(theta, r0, r1)
	//	slope_x, slope_y := float32(slope_x_), float32(slope_y_)
	if math.IsNaN(slope_x) {
		log.Printf("slope_x %v %v %v", slope_x, theta, phi)
	}
	if math.IsNaN(slope_y) {
		log.Printf("slope_y %v %v %v", slope_y, theta, phi)
	}

	tmp := math.Cos(phi)*slope_x - math.Sin(phi)*slope_y
	slope_y = math.Sin(phi)*slope_x + math.Cos(phi)*slope_y
	slope_x = tmp

	slope_x = alpha_x * slope_x
	slope_y = alpha_y * slope_y

	inv_omega_m := math.Sqrt(slope_x*slope_x + slope_y*slope_y + 1.0)
	omega_m[0] = -float32(slope_x / inv_omega_m)
	omega_m[1] = -float32(slope_y / inv_omega_m)
	omega_m[2] = float32(1.0 / inv_omega_m)
	return
}

func (b *CookTorranceGGX2) sample11(theta_i, r0, r1 float64) (slope_x, slope_y float64) {
	if theta_i < 0.0001 {
		r := math.Sqrt(r0 / (1 - r0))
		phi := 6.28318530718 * r1
		slope_x = r * math.Cos(phi)
		slope_y = r * math.Sin(phi)
		return
	}
	// precomputations
	tan_theta_i := math.Tan(theta_i)
	a := 1.0 / (tan_theta_i)
	G1 := 2.0 / (1 + math.Sqrt(1.0+1.0/(a*a)))

	// sample slope_x
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
	B := tan_theta_i
	D := math.Sqrt(B*B*tmp*tmp - (A*A-B*B)*tmp)

	slope_x_1 := B*tmp - D
	slope_x_2 := B*tmp + D

	if A < 0 || slope_x_2 > 1.0/tan_theta_i {
		slope_x = slope_x_1
	} else {
		slope_x = slope_x_2
	}

	if math.IsNaN(slope_x_1) || math.IsNaN(slope_x_2) {
		log.Printf("slope %v %v tan_theta:%v B:%v tmp:%v D:%v A:%v G1:%v a:%v r0:%v r1:%v", slope_x_1, slope_x_2, tan_theta_i, B, tmp, D, A, G1, a, r0, r1)
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

	slope_y = S * z * math.Sqrt(1.0+slope_x*slope_x)

	return
}
