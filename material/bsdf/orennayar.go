// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsdf

import (
	"errors"
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/math/sample"
	"math"
)

var ErrFailedEval = errors.New("Failed Eval")

// Instanced for each point
type OrenNayar2 struct {
	Lambda    float32
	OmegaI    m.Vec3
	Roughness float32
}

func NewOrenNayar(sg *core.ShaderGlobals, roughness float32) *OrenNayar2 {
	return &OrenNayar2{sg.Lambda, sg.ViewDirection(), roughness * roughness}
}

func (b *OrenNayar2) Sample(r0, r1 float64) m.Vec3 {
	return sample.CosineHemisphere(r0, r1)
}

func (b *OrenNayar2) PDF(omega_o m.Vec3) float64 {
	o_dot_n := float64(m.Abs(omega_o[2]))

	return o_dot_n / math.Pi
}

func (b *OrenNayar2) Eval(omega_o m.Vec3) (rho colour.Spectrum) {

	sigma := b.Roughness

	A := 1 - (0.5 * (sigma * sigma) / ((sigma * sigma) + 0.57))

	B := 0.45 * (sigma * sigma) / ((sigma * sigma) + 0.09)

	phi_i := m.Atan2(b.OmegaI[1], b.OmegaI[0])
	phi_o := m.Atan2(omega_o[1], omega_o[0])
	theta_i := m.Acos(b.OmegaI[2])
	theta_o := m.Acos(omega_o[2])

	alpha := m.Max(theta_i, theta_o)
	beta := m.Min(theta_i, theta_o)

	C := m.Sin(alpha) * m.Tan(beta)

	gamma := m.Cos(phi_o - phi_i)
	// This might be ok:
	// gamma := dot(eyeDir - normal * dot(eyeDir, normal), light.direction - normal * dot(light.direction, normal))

	//*out = b.Kd
	scale := omega_o[2] * (A + (B * m.Max(0, gamma) * C))

	rho.Lambda = b.Lambda
	rho.FromRGB(1, 1, 1)
	rho.Scale(scale / math.Pi)

	return
}
