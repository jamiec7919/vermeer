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
	"math/rand"
)

// Weight = f * dot(omega_o,n) / pdf
// pdf := dot(omega_o,n) / m.Pi
// f := rho / m.Pi
type Diffuse struct {
	Kd core.MapSampler
}

func (b *Diffuse) IsDelta(shade *core.SurfacePoint) bool { return false }

func (b *Diffuse) ContinuationProb(shade *core.SurfacePoint) float64 {
	return 1.0
}

func (b *Diffuse) PDF(shade *core.SurfacePoint, omega_i, omega_o m.Vec3) float64 {
	o_dot_n := float64(omega_o[2])

	//if o_dot_n < 0.0 {
	//	log.Printf("BSDFDiffuse.PDF %v", o_dot_n)
	//}
	// Sampling is cosine weighted so by projected-solid-angle this is:

	return o_dot_n / math.Pi
}

func (b *Diffuse) Sample(shade *core.SurfacePoint, omega_i m.Vec3, rnd *rand.Rand, omega_o *m.Vec3, rho *colour.Spectrum, pdf *float64) error {
	*omega_o = sample.CosineHemisphere(rnd.Float64(), rnd.Float64())

	*pdf = b.PDF(shade, omega_i, *omega_o)

	col := b.Kd.SampleRGB(shade.UV[0][0], shade.UV[0][1], 1, 1)
	//colour.RGBToSpectrumSmits99(col[0], col[1], col[2], out)

	rho.FromRGB(col[0], col[1], col[2])
	//*out = b.Kd
	rho.Scale(1 / math.Pi)
	//b.Diffuse.SampleRGB(float32(shade.UVSet[0].uv[0]), float32(shade.UVSet[0].uv[1]), shade.UVSet[0].dTd0, shade.UVSet[0].dTd1, 1, 1)
	return nil
}

func (b *Diffuse) Eval(shade *core.SurfacePoint, omega_i, omega_o m.Vec3, rho *colour.Spectrum) error {
	o_dot_n := omega_o[2]

	if o_dot_n <= 0 {
		//out.C[1] = 1
		//o_dot_n = -o_dot_n
		//panic("bsdf.Diffuse.Eval")
		return nil
	}
	//ColourScale(1.0*o_dot_n/math.Pi, b.Diffuse.SampleRGB(float32(shade.UVSet[0].uv[0]), float32(shade.UVSet[0].uv[1]), shade.UVSet[0].dTd0, shade.UVSet[0].dTd1, 1, 1))
	Kd := b.Kd.SampleRGB(shade.UV[0][0], shade.UV[0][1], 1, 1)
	//colour.RGBToSpectrumSmits99(col[0], col[1], col[2], out)

	rho.FromRGB(Kd[0], Kd[1], Kd[2])
	//*out = b.Kd
	rho.Scale(1 / math.Pi)
	return nil
}
