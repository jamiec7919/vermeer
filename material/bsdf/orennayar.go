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
	"math/rand"
)

var ErrFailedEval = errors.New("Failed Eval")

type OrenNayar struct {
	Kd        core.MapSampler
	Roughness core.MapSampler
}

func (b *OrenNayar) IsDelta(shade *core.SurfacePoint) bool { return false }

func (b *OrenNayar) ContinuationProb(shade *core.SurfacePoint) float64 {
	return 1.0
}

func (b *OrenNayar) PDF(shade *core.SurfacePoint, omega_i, omega_o m.Vec3) float64 {
	o_dot_n := float64(omega_o[2])

	//if o_dot_n < 0.0 {
	//	log.Printf("BSDFDiffuse.PDF %v", o_dot_n)
	//}
	// Sampling is cosine weighted so by projected-solid-angle this is:

	return o_dot_n / math.Pi
}

func (b *OrenNayar) Sample(shade *core.SurfacePoint, omega_i m.Vec3, rnd *rand.Rand, omega_o *m.Vec3, rho *colour.Spectrum, pdf *float64) error {
	*omega_o = sample.CosineHemisphere(rnd.Float64(), rnd.Float64())

	*pdf = b.PDF(shade, omega_i, *omega_o)

	return b.Eval(shade, omega_i, *omega_o, rho)
}

func (b *OrenNayar) Eval(shade *core.SurfacePoint, omega_i, omega_o m.Vec3, rho *colour.Spectrum) error {
	o_dot_n := omega_o[2]

	if o_dot_n <= 0 {
		//out.C[1] = 1
		//o_dot_n = -o_dot_n
		//panic("bsdf.OrenNayar.Eval")
		return ErrFailedEval
	}

	sigma := b.Roughness.SampleScalar(shade.UV[0][0], shade.UV[0][1], 1, 1)
	//ColourScale(1.`0*o_dot_n/math.Pi, b.Diffuse.SampleRGB(float32(shade.UVSet[0].uv[0]), float32(shade.UVSet[0].uv[1]), shade.UVSet[0].dTd0, shade.UVSet[0].dTd1, 1, 1))
	Kd := b.Kd.SampleRGB(shade.UV[0][0], shade.UV[0][1], 1, 1)
	//colour.RGBToSpectrumSmits99(col[0], col[1], col[2], out)
	A := 1 - (0.5 * (sigma * sigma) / ((sigma * sigma) + 0.57))

	B := 0.45 * (sigma * sigma) / ((sigma * sigma) + 0.09)

	phi_i := m.Atan2(omega_i[1], omega_i[0])
	phi_o := m.Atan2(omega_o[1], omega_o[0])
	theta_i := m.Acos(omega_i[2])
	theta_o := m.Acos(omega_o[2])

	alpha := m.Max(theta_i, theta_o)
	beta := m.Min(theta_i, theta_o)

	rho.FromRGB(Kd[0], Kd[1], Kd[2])

	C := m.Sin(alpha) * m.Tan(beta)

	gamma := m.Cos(phi_i - phi_o)
	// This might be ok:
	// gamma := dot(eyeDir - normal * dot(eyeDir, normal), light.direction - normal * dot(light.direction, normal))

	//*out = b.Kd
	scale := omega_i[2] * (A + (B * m.Max(0, gamma) * C))

	rho.Scale(scale / math.Pi)
	return nil
}
