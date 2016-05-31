// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package material

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/material/bsdf"
	mt "github.com/jamiec7919/vermeer/math"
	//"log"
)

// Eval implements core.Material.  Performs all shading for the surface point in sg.  May trace
// rays and shadow rays.
func (m *Material) Eval(sg *core.ShaderGlobals) {
	if sg.Depth > 3 {
		return
	}

	if m.HasBumpMap() {
		m.ApplyBumpMap(sg)
	}

	roughness := float32(0.5)
	if m.Roughness != nil {
		roughness = m.Roughness.Float32(sg)
	}

	specRoughness := float32(0.5)
	if m.SpecularRoughness != nil {
		specRoughness = m.SpecularRoughness.Float32(sg)
	}
	ior := float32(1.7)
	if m.IOR != nil {
		ior = m.IOR.Float32(sg)
	}
	var brdf2 core.BRDF

	if specRoughness == 0.0 {
		brdf2 = bsdf.NewSpecular(sg)
	} else {
		brdf2 = bsdf.NewMicrofacetGGX(sg, ior, specRoughness)
	}

	brdf := bsdf.NewOrenNayar(sg, roughness)
	//brdf := bsdf.NewLambert(sg)
	sg.LightsPrepare()

	var diffcontrib colour.RGB

	Kd := m.Kd.RGB(sg)

	for sg.LightsGetSample() {

		if sg.Lp.DiffuseShadeMult() > 0.0 {

			// In this example the brdf passed is an interface
			// allowing sampling, pdf and bsdf eval
			col := sg.EvaluateLightSample(brdf)

			col.Mul(Kd)
			diffcontrib.Add(col)
		}

	}
	/*
		if m.Ks != nil {
			spec := m.Ks.RGB(sg)

			if spec[0] > 0.0 || spec[1] > 0.0 || spec[2] > 0.0 {
				diffcontrib.Mul(spec)
			}

		}
	*/
	var speccontrib colour.RGB

	if m.Ks != nil {
		var samp core.ScreenSample
		ray := new(core.RayData)
		s := sg.GlossySample(brdf2)

		if mt.Vec3Length(s) < 0.001 {
			goto skip
		}
		sg.Depth++
		ray.Init(0, sg.OffsetP(1), sg.TangentToWorld(s), mt.Inf(1), sg)

		if core.Trace(ray, &samp) {
			rho := brdf2.Eval(s)
			rho.Scale(sg.Weight)
			r, g, b := rho.ToRGB()
			specrgb := colour.RGB(m.Ks.RGB(sg))
			specrgb.Mul(colour.RGB{r, g, b})
			specrgb.Mul(samp.Colour)

			speccontrib.Add(specrgb)
		} else {
			/*
				rho := brdf2.Eval(s)
				rho.Scale(sg.Weight)
				r, g, b := rho.ToRGB()
				specrgb := colour.RGB(m.Ks.RGB(sg))
				specrgb.Mul(colour.RGB{r, g, b})
				specrgb.Mul(colour.RGB{10, 10, 10})
				speccontrib.Add(specrgb)

				rhod := brdf.Eval(s)
				r, g, b = rhod.ToRGB()
				diffrgb := colour.RGB{r, g, b}
				diffrgb.Mul(Kd)
				diffrgb.Mul(colour.RGB{10, 10, 10})
				diffcontrib.Add(diffrgb)
			*/
		}

		speccontrib.Scale(0.5)
		diffcontrib.Scale(0.5)
		sg.OutRGB.Add(speccontrib)
	}
skip:
	sg.OutRGB.Add(diffcontrib)

	if m.E != nil {
		E := m.E.RGB(sg)
		sg.OutRGB.Add(E)
	}

}
