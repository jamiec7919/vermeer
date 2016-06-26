// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package bsdf provides built-in B(R/S)DF (Bidirectional Reflectance/Scattering Distribution Function)
models for Vermeer. */
package bsdf

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/math/sample"
	"math"
)

// Weight = f * dot(omegaO,n) / pdf
// pdf := dot(omegaO,n) / m.Pi
// f := rho / m.Pi

// Instanced for each point

// Lambert is basic lambertian diffuse.
type Lambert struct {
	Lambda float32
	OmegaI m.Vec3
}

// NewLambert will return an instance of the model for the given point.
func NewLambert(sg *core.ShaderGlobals) *Lambert {
	return &Lambert{sg.Lambda, sg.ViewDirection()}
}

// Sample implements core.BSDF.
func (b *Lambert) Sample(r0, r1 float64) m.Vec3 {
	return sample.CosineHemisphere(r0, r1)
}

// PDF implements core.BSDF.
func (b *Lambert) PDF(omegaO m.Vec3) float64 {
	ODotN := float64(m.Abs(omegaO[2]))

	return ODotN / math.Pi
}

// Eval implements core.BSDF.
func (b *Lambert) Eval(omegaO m.Vec3) (rho colour.Spectrum) {
	weight := omegaO[2]

	rho.Lambda = b.Lambda
	rho.FromRGB(1, 1, 1)
	rho.Scale(weight / m.Pi)

	return
}
