// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsdf

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
)

func reflect(omega_in, N m.Vec3) (omega_out m.Vec3) {
	omega_out = m.Vec3Add(m.Vec3Neg(omega_in), m.Vec3Scale(2*m.Vec3Dot(omega_in, N), N))
	return
}

// Specular2 implements perfect mirror specular reflection.
// Instanced for each point
type Specular2 struct {
	Lambda float32
	OmegaI m.Vec3
}

// NewSpecular returns a new instance of the model.
func NewSpecular(sg *core.ShaderGlobals) *Specular2 {
	return &Specular2{sg.Lambda, sg.ViewDirection()}
}

// Sample implements core.BSDF.
func (b *Specular2) Sample(r0, r1 float64) m.Vec3 {
	omega_i := b.OmegaI

	if omega_i[2] < 0 {
		//log.Printf("Specular.Sample: %v", omega_i)
	}
	//out.Omega = reflect(shade.Omega, m.Vec3{0, 0, 1})
	omega_o := reflect(omega_i, m.Vec3{0, 0, 1})

	if d := m.Vec3Dot(omega_o, m.Vec3{0, 0, 1}); d <= 0.0 {
		return reflect(omega_i, m.Vec3{0, 0, -1})
		//log.Printf("Err dot %v", d)
		//	*omega_o = m.Vec3Add(m.Vec3Neg(*omega_o), m.Vec3Scale(2*m.Vec3Dot(*omega_o, m.Vec3{0, 0, 1}), m.Vec3{0, 0, 1}))
	}
	return omega_o
}

// PDF implements core.BSDF.
func (b *Specular2) PDF(omega_o m.Vec3) float64 {
	return 1
}

// Eval implements core.BSDF.
func (b *Specular2) Eval(omega_o m.Vec3) (rho colour.Spectrum) {
	rho.Lambda = b.Lambda
	rho.FromRGB(1, 1, 1)
	return
}
