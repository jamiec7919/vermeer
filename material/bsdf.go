// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package material

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/material/bsdf"
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

type BSDF interface {

	// Compute the PDF for the given omega in (shade) and out vector omega_o
	PDF(shade *core.SurfacePoint, omega_i, omega_o m.Vec3) float64

	// Sample an outbound direction for the incoming direction in shade.Omega.
	// Return core.ErrNoSample if unable to compute a sample for the given configuration
	Sample(shade *core.SurfacePoint, omega_i m.Vec3, rnd *rand.Rand, omega_o *m.Vec3, rho *colour.Spectrum, pdf *float64) error

	// Evaluate the BSDF function for the given parameters.  Spectrum is stored in out.
	// Does not divide by the PDF, this should be computed seperately.
	Eval(shade *core.SurfacePoint, omega_i, omega_o m.Vec3, rho *colour.Spectrum) error

	// Returns true if function includes a dirac delta (e.g. specular reflection)
	IsDelta(shade *core.SurfacePoint) bool

	// Probability that the path with the shade point as the end should continue.
	ContinuationProb(shade *core.SurfacePoint) float64
}

type LayeredBSDF struct {
	Layers  []BSDF
	Weights []float64
}

/*
type WeightedBSDF struct {
	BSDF    []BSDF
	Weights []ScalarSampler
}

func (b *WeightedBSDF) PDF(shade *ShadePoint, omega_o m.Vec3) float64 {
	return 0
}

func (b *WeightedBSDF) Sample(shade *ShadePoint, rnd *rand.Rand, out *DirectionalSample) error {
	return nil
}

func (b *WeightedBSDF) Eval(shade *ShadePoint, omega_o m.Vec3, out *colour.Spectrum) error { return nil }
*/

func makeBSDF(mtl *Material) BSDF {
	ty := mtl.Diffuse

	switch ty {
	case "Lambert":
		bsdf := bsdf.Diffuse{Kd: mtl.Kd}
		return &bsdf
	case "OrenNayar":
		bsdf := bsdf.OrenNayar{Roughness: mtl.Roughness, Kd: mtl.Kd}
		return &bsdf
	}

	ty = mtl.Specular

	switch ty {
	case "Specular":
		bsdf := bsdf.Specular{Ks: mtl.Ks}

		return &bsdf

	case "GGXSpecular":
		bsdf := bsdf.CookTorranceGGX{Roughness: mtl.Roughness, Ks: mtl.Ks, IOR: mtl.IOR}
		return &bsdf
	}

	return nil
}
