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

func makeBSDF(params core.Params) BSDF {
	ty, _ := params.GetString("Type")

	switch ty {
	case "Lambert":
		bsdf := bsdf.Diffuse{}
		Kd := params["Kd"]

		switch t := Kd.(type) {
		case string:
			bsdf.Kd = &TextureMap{t}
		case []float64:
			bsdf.Kd = &ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		case []int64:
			bsdf.Kd = &ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		}
		return &bsdf
	case "OrenNayar":
		bsdf := bsdf.OrenNayar{}

		roughness := params["Roughness"]

		switch t := roughness.(type) {
		case string:
			bsdf.Roughness = &TextureMap{t}
		case float64:
			bsdf.Roughness = &ConstantMap{[3]float32{float32(t)}}
		case int64:
			bsdf.Roughness = &ConstantMap{[3]float32{float32(t)}}
		}

		Kd := params["Kd"]

		switch t := Kd.(type) {
		case string:
			bsdf.Kd = &TextureMap{t}
		case []float64:
			bsdf.Kd = &ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		case []int64:
			bsdf.Kd = &ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		}
		return &bsdf

	case "Specular":
		bsdf := bsdf.Specular{}

		Ks, present := params["Ks"]

		if !present {
			Ks = []float64{0.5, 0.5, 0.5}
		}

		switch t := Ks.(type) {
		case string:
			bsdf.Ks = &TextureMap{t}
		case []float64:
			bsdf.Ks = &ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		case []int64:
			bsdf.Ks = &ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		}
		return &bsdf

	case "GGXSpecular":
		bsdf := bsdf.CookTorranceGGX{}

		roughness := params["Roughness"]

		switch t := roughness.(type) {
		case string:
			bsdf.Roughness = &TextureMap{t}
		case float64:
			bsdf.Roughness = &ConstantMap{[3]float32{float32(t)}}
		case int64:
			bsdf.Roughness = &ConstantMap{[3]float32{float32(t)}}
		}

		ior := params["IOR"]

		switch t := ior.(type) {
		case string:
			bsdf.IOR = &TextureMap{t}
		case float64:
			bsdf.IOR = &ConstantMap{[3]float32{float32(t)}}
		case int64:
			bsdf.IOR = &ConstantMap{[3]float32{float32(t)}}
		}

		Ks, present := params["Ks"]

		if !present {
			Ks = []float64{0.5, 0.5, 0.5}
		}

		switch t := Ks.(type) {
		case string:
			bsdf.Ks = &TextureMap{t}
		case []float64:
			bsdf.Ks = &ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		case []int64:
			bsdf.Ks = &ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		}
		return &bsdf
	}

	return nil
}
