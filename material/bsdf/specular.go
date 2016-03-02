// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bsdf

import (
	"github.com/jamiec7919/vermeer/material"
	m "github.com/jamiec7919/vermeer/math"
	"log"
	"math/rand"
)

type Specular struct {
	Ks material.MapSampler
}

func (b *Specular) IsDelta(shade *material.SurfacePoint) bool { return true }

func (b *Specular) ContinuationProb(shade *material.SurfacePoint) float64 {
	return 1.0
}

func reflect(omega_in, N m.Vec3) (omega_out m.Vec3) {
	omega_out = m.Vec3Add(m.Vec3Neg(omega_in), m.Vec3Scale(2*m.Vec3Dot(omega_in, N), N))
	return
}

func (b *Specular) PDF(shade *material.SurfacePoint, omega_i, omega_o m.Vec3) float64 {
	//o_dot_n := float64(omega_o[2])

	//if o_dot_n < 0.0 {
	//	log.Printf("BSDFDiffuse.PDF %v", o_dot_n)
	//}
	// Sampling is cosine weighted so by projected-solid-angle this is:
	//R := reflect(shade.Omega, m.Vec3{0, 0, 1})

	R := reflect(omega_i, shade.Ns)
	if d := m.Vec3Dot(R, m.Vec3{0, 0, 1}); d <= 0.0 {
		//log.Printf("Err dot %v", d)
		R = m.Vec3Neg(R)
	}
	//d := m.Vec3Dot(R, omega_o)
	//	if d < 1.0-0.99999 || d > 1.0+0.99999 {
	//		return 0
	//	}

	return 1
}

func (b *Specular) Sample(shade *material.SurfacePoint, omega_i m.Vec3, rnd *rand.Rand, omega_o *m.Vec3, rho *material.Spectrum, pdf *float64) error {

	if omega_i[2] < 0 {
		log.Printf("ERR: %v", omega_i)
	}
	//out.Omega = reflect(shade.Omega, m.Vec3{0, 0, 1})
	*omega_o = reflect(omega_i, shade.Ns)

	if d := m.Vec3Dot(*omega_o, m.Vec3{0, 0, 1}); d <= 0.0 {
		//log.Printf("Err dot %v", d)
		*omega_o = m.Vec3Add(m.Vec3Neg(*omega_o), m.Vec3Scale(2*m.Vec3Dot(*omega_o, m.Vec3{0, 0, 1}), m.Vec3{0, 0, 1}))
	}
	*pdf = b.PDF(shade, omega_i, *omega_o)

	col := b.Ks.SampleRGB(shade.UV[0][0], shade.UV[0][1], 1, 1)
	//colour.RGBToSpectrumSmits99(col[0], col[1], col[2], out)

	rho.FromRGB(col[0], col[1], col[2])
	//*out = b.Kd
	//b.Diffuse.SampleRGB(float32(shade.UVSet[0].uv[0]), float32(shade.UVSet[0].uv[1]), shade.UVSet[0].dTd0, shade.UVSet[0].dTd1, 1, 1)
	return nil
}

func (b *Specular) Eval(shade *material.SurfacePoint, omega_i, omega_o m.Vec3, rho *material.Spectrum) error {
	//R := reflect(shade.Omega, m.Vec3{0, 0, 1})
	//R := reflect(omega_i, shade.Ns)

	//if d := m.Vec3Dot(R, omega_o); d != 1.0 {
	//log.Printf("Err dot %v", d)
	//	R = m.Vec3Neg(R)
	//	}
	//d := m.Vec3Dot(R, omega_o)
	//	if d < 1.0-0.99999 || d > 1.0+0.99999 {
	//		out.SetZero()
	//		return nil
	//	}
	//ColourScale(1.0*o_dot_n/math.Pi, b.Diffuse.SampleRGB(float32(shade.UVSet[0].uv[0]), float32(shade.UVSet[0].uv[1]), shade.UVSet[0].dTd0, shade.UVSet[0].dTd1, 1, 1))
	col := b.Ks.SampleRGB(shade.UV[0][0], shade.UV[0][1], 1, 1)
	//colour.RGBToSpectrumSmits99(col[0], col[1], col[2], out)

	rho.FromRGB(col[0], col[1], col[2])
	return nil
}
