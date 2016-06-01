// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package material

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/material/bsdf"
	fr "github.com/jamiec7919/vermeer/material/fresnel"
	m "github.com/jamiec7919/vermeer/math"
	"log"
)

// Eval implements core.Material.  Performs all shading for the surface point in sg.  May trace
// rays and shadow rays.
func (mtl *Material) Eval(sg *core.ShaderGlobals) {
	if sg.Depth > 4 {
		return
	}

	if mtl.HasBumpMap() {
		mtl.ApplyBumpMap(sg)
	} else {
		sg.N = m.Vec3Normalize(sg.N)
	}

	roughness := float32(0.5)
	if mtl.Roughness != nil {
		roughness = mtl.Roughness.Float32(sg)
	}

	specRoughness := float32(0.5)
	if mtl.SpecularRoughness != nil {
		specRoughness = mtl.SpecularRoughness.Float32(sg)
	}

	ior := float32(1.7)

	if mtl.IOR != nil {
		ior = mtl.IOR.Float32(sg)
	}

	var brdf2, btdf core.BSDF

	transWeight := float32(0)

	if mtl.TransStrength != nil {
		transWeight = mtl.TransStrength.Float32(sg)
	}

	transmissive := false

	var fresnel core.Fresnel

	switch mtl.spec1FresnelModel {
	case FRESNEL_DIELECTRIC:
		fresnel = fr.NewDielectric(ior)
	case FRESNEL_METAL:

		refl := colour.RGB{0.5, 0.5, 0.5}
		edge := colour.RGB{0.5, 0.5, 0.5}

		if mtl.Spec1FresnelRefl != nil {
			refl = mtl.Spec1FresnelRefl.RGB(sg)
		}

		if mtl.Spec1FresnelEdge != nil {
			refl = mtl.Spec1FresnelEdge.RGB(sg)
		}

		fresnel = fr.NewConductor(0, refl, edge)
	}

	if transWeight > 0.0 {
		transmissive = true

		if specRoughness == 0.0 {
			btdf = bsdf.NewSpecularTransmission(sg, ior, fresnel, mtl.TransThin)

		} else {
			btdf = bsdf.NewMicrofacetTransmissionGGX(sg, ior, specRoughness, fresnel, mtl.TransThin)
		}
	}

	diffWeight := float32(0.5)
	specWeight := float32(0.5)
	if mtl.DiffuseStrength != nil {
		diffWeight = mtl.DiffuseStrength.Float32(sg)
	}

	if mtl.SpecularStrength != nil {
		specWeight = mtl.SpecularStrength.Float32(sg)
	}

	// Normalize the weights
	l := diffWeight*diffWeight + specWeight*specWeight
	l = m.Sqrt(l)
	if l != 0.0 {
		diffWeight /= l
		specWeight /= l

	}

	Ks := colour.RGB{}
	if mtl.Ks != nil {
		Ks = mtl.Ks.RGB(sg)
	}

	Kt := colour.RGB{}
	if mtl.Kt != nil {
		Kt = mtl.Kt.RGB(sg)
	}

	if specRoughness == 0.0 {
		brdf2 = bsdf.NewSpecular(sg, fresnel, transWeight, mtl.TransThin)
	} else {
		brdf2 = bsdf.NewMicrofacetGGX(sg, fresnel, specRoughness, transmissive, mtl.TransThin)
	}

	brdf := bsdf.NewOrenNayar(sg, roughness)
	//brdf := bsdf.NewLambert(sg)

	var diffcontrib colour.RGB

	if diffWeight > 0.0 {
		sg.LightsPrepare()

		Kd := mtl.Kd.RGB(sg)

		for sg.LightsGetSample() {

			if sg.Lp.DiffuseShadeMult() > 0.0 {

				// In this example the brdf passed is an interface
				// allowing sampling, pdf and bsdf eval
				col := sg.EvaluateLightSample(brdf)
				col.Mul(Kd)
				diffcontrib.Add(col)
			}

		}
	}

	/*
		if mtl.Ks != nil {
			spec := mtl.Ks.RGB(sg)

			if spec[0] > 0.0 || spec[1] > 0.0 || spec[2] > 0.0 {
				diffcontrib.Mul(spec)
			}

		}
	*/
	var speccontrib colour.RGB

	if mtl.Ks != nil {
		var samp core.ScreenSample
		ray := new(core.RayData)

		s := m.Vec3{}

		r0 := sg.Rand().Float64()

		frrgb := fresnel.Kr(m.Vec3DotAbs(sg.ViewDirection(), m.Vec3{0, 0, 1}))
		transmit := false

		pdf := 1.0

		if !transmissive || r0 < float64(frrgb.Maxh()) {
			s = sg.GlossySample(brdf2)

			if transmissive {
				pdf = float64(frrgb.Maxh())
			}
		} else {
			s = sg.GlossySample(btdf)
			transmit = true
			pdf = 1.0 - float64(frrgb.Maxh())
		}

		if m.Vec3Length(s) < 0.9 {
			log.Printf("err %v %v", m.Vec3Length(s), s)
			goto skip
		}
		sg.Depth++

		// this is wrong, classifies transmitted rays correctly but not acute reflected!!
		//if m.Vec3Dot(s, sg.ViewDirection()) < 0
		//log.Printf("%v %v", s, sg.ViewDirection())
		if m.Vec3Dot(sg.TangentToWorld(s), sg.Ng) < 0 {
			//if m.Vec3Dot(sg.TangentToWorld(s), sg.N) < 0 && m.Vec3Dot(m.Vec3Neg(sg.Rd), sg.N) > 0 {
			ray.Init(0, sg.OffsetP(-1), sg.TangentToWorld(s), m.Inf(1), sg)
		} else {
			ray.Init(0, sg.OffsetP(1), sg.TangentToWorld(s), m.Inf(1), sg)

		}

		if core.Trace(ray, &samp) {

			if !transmit {
				rho := brdf2.Eval(s)

				if sg.Weight < 1000000 {
					rho.Scale(sg.Weight / float32(pdf))

					//log.Printf("%v %v", sg.Weight, rho)
					r, g, b := rho.ToRGB()
					specrgb := colour.RGB(Ks)
					specrgb.Mul(colour.RGB{r, g, b})
					specrgb.Mul(samp.Colour)

					if transWeight == 0 {
						rho := brdf.Eval(s)
						//				rho.Scale(0.9) // should be 1-Fresnel
						r, g, b := rho.ToRGB()
						diffrgb := mtl.Kd.RGB(sg)
						diffrgb.Mul(colour.RGB{r, g, b})
						diffrgb.Mul(samp.Colour)
						//			diffrgb.Scale(100)
						//specrgb.Mul(diffrgb)
						diffcontrib.Add(diffrgb)
					}
					speccontrib.Add(specrgb)
				}
			} else {
				rho := btdf.Eval(s)

				rho.Scale(sg.Weight / float32(pdf))
				r, g, b := rho.ToRGB()
				specrgb := colour.RGB(Kt)
				specrgb.Mul(colour.RGB{r, g, b})
				specrgb.Mul(samp.Colour)
				speccontrib.Add(specrgb)

			}

		} else {
			/*
				rho := brdf2.Eval(s)
				rho.Scale(sg.Weight)
				r, g, b := rho.ToRGB()
				specrgb := colour.RGB(mtl.Ks.RGB(sg))
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

		speccontrib.Scale(specWeight)
		diffcontrib.Scale(diffWeight)
		sg.OutRGB.Add(speccontrib)
	}
skip:
	sg.OutRGB.Add(diffcontrib)

	if mtl.E != nil {
		E := mtl.E.RGB(sg)
		sg.OutRGB.Add(E)
	}

}
