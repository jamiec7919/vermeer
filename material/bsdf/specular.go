// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsdf

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"log"
)

func reflect(omegaR, N m.Vec3) (omegaO m.Vec3) {
	//omegaO = m.Vec3Add(m.Vec3Neg(omegaR), m.Vec3Scale(2*m.Vec3DotAbs(omegaR, N), N))

	omegaO = m.Vec3Sub(m.Vec3Scale(2.0*m.Vec3Dot(N, omegaR), N), omegaR)

	return
}

func refract(omegaR m.Vec3, ior float32) (omegaO m.Vec3) {
	eta := 1.0 / ior
	sign := sign(omegaR[2])

	if sign < 0 {
		// Approaching from 'in media' so swap etaI & etaO
		eta = ior
	}

	c := m.Vec3Dot(omegaR, m.Vec3{0, 0, 1})

	a := eta*c - sign*m.Sqrt(1.0+eta*(c*c-1.0))

	omegaO = m.Vec3Sub(m.Vec3Scale(a, m.Vec3{0, 0, 1}), m.Vec3Scale(eta, omegaR))

	return
}

// Specular2 implements perfect mirror specular reflection.
// Instanced for each point
type Specular2 struct {
	Lambda       float32
	OmegaR       m.Vec3
	fresnel      core.Fresnel
	transmissive float32
	thin         bool
}

// NewSpecular returns a new instance of the model.
func NewSpecular(sg *core.ShaderGlobals, fresnel core.Fresnel, transmissive float32, thin bool) *Specular2 {
	return &Specular2{sg.Lambda, sg.ViewDirection(), fresnel, transmissive, thin}
}

// Sample implements core.BSDF.
func (b *Specular2) Sample(r0, r1 float64) (omegaO m.Vec3) {

	//fresnel := fresnel.Kr(m.Vec3DotAbs(b.OmegaRm.Vec3{0, 0, 1})).Maxh()

	//if b.transmissive == 0.0 || r0 < float64(fresnel) {
	omegaO = reflect(b.OmegaR, m.Vec3{0, 0, 1})

	if sign(omegaO[2]) != sign(b.OmegaR[2]) {
		log.Panicf("signs %v %v", omegaO, b.OmegaR)
	}
	//	} else {
	//		if b.thin {
	//			omegaO = m.Vec3Neg(b.OmegaR)
	//		} else {
	//		log.Printf("transmit")
	//omegaO = refract(b.OmegaR, b.ior)
	//		}
	//	}

	omegaO = m.Vec3Normalize(omegaO)
	return
}

// PDF implements core.BSDF.
func (b *Specular2) PDF(omegaO m.Vec3) float64 {
	return 1
}

// Eval implements core.BSDF.
func (b *Specular2) Eval(omegaO m.Vec3) (rho colour.Spectrum) {
	fresnel := b.fresnel.Kr(b.OmegaR[2])

	rho.Lambda = b.Lambda
	rho.FromRGB(fresnel[0], fresnel[1], fresnel[2])
	rho.Scale(m.Vec3DotAbs(omegaO, m.Vec3{0, 0, 1}))
	return
}
