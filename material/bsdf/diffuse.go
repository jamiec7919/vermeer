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

// Weight = f * dot(omega_o,n) / pdf
// pdf := dot(omega_o,n) / m.Pi
// f := rho / m.Pi

// Instanced for each point
type Lambert struct {
	Lambda float32
	OmegaI m.Vec3
}

func NewLambert(sg *core.ShaderGlobals) *Lambert {
	return &Lambert{sg.Lambda, sg.ViewDirection()}
}

func (b *Lambert) Sample(r0, r1 float64) m.Vec3 {
	return sample.CosineHemisphere(r0, r1)
}

func (b *Lambert) PDF(omega_o m.Vec3) float64 {
	o_dot_n := float64(m.Abs(omega_o[2]))

	return o_dot_n / math.Pi
}

func (b *Lambert) Eval(omega_o m.Vec3) (rho colour.Spectrum) {
	weight := omega_o[2]

	rho.Lambda = b.Lambda
	rho.FromRGB(1, 1, 1)
	rho.Scale(weight / m.Pi)

	return
}
