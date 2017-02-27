// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package shader provides the default shader(s) for Vermeer.

This package is in heavy development so documentation somewhat sketchy. */
package shader

import (
	"fmt"
	"github.com/jamiec7919/vermeer/builtin/shader/bsdf"
	fr "github.com/jamiec7919/vermeer/builtin/shader/fresnel"
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/core/param"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/math/ldseq"
	"github.com/jamiec7919/vermeer/nodes"
	"math"
)

// ShaderStd is the default surface shader.
type ShaderStd struct {
	NodeDef core.NodeDef `node:"-"`
	MtlName string       `node:"Name"`

	EmissionColour   param.RGBUniform     `node:",opt"`
	EmissionStrength param.Float32Uniform `node:",opt"`

	Sides int `node:",opt"` // One or two sided

	DiffuseColour    param.RGBUniform     `node:",opt"` // Colour parameter
	DiffuseStrength  param.Float32Uniform `node:",opt"` // Weight parameter
	DiffuseRoughness param.Float32Uniform `node:",opt"` // Oren-Nayar Roughness parameter

	Spec1Colour       param.RGBUniform     `node:",opt"` // Colour parameter
	Spec1Strength     param.Float32Uniform `node:",opt"` // Weight parameter
	Spec1Roughness    param.Float32Uniform `node:",opt"`
	Spec1FresnelModel string               `node:",opt"`
	spec1FresnelModel fr.Model
	Spec1FresnelRefl  param.RGBUniform `node:",opt"` // Colour parameter
	Spec1FresnelEdge  param.RGBUniform `node:",opt"` // Colour parameter

	IOR param.Float32Uniform `node:",opt"`
}

// Assert that ShaderStd satisfies important interfaces.
var _ core.Node = (*ShaderStd)(nil)
var _ core.Shader = (*ShaderStd)(nil)

// Name is a core.Node method.
func (sh *ShaderStd) Name() string { return sh.MtlName }

// Def is a core.Node method.
func (sh *ShaderStd) Def() core.NodeDef { return sh.NodeDef }

// PreRender is a core.Node method.
func (sh *ShaderStd) PreRender() error {

	switch sh.Spec1FresnelModel {
	case "Dielectric":
		sh.spec1FresnelModel = fr.DielectricModel
	case "Metal":
		sh.spec1FresnelModel = fr.ConductorModel

	}
	return nil
}

// PostRender is a core.Node method.
func (sh *ShaderStd) PostRender() error { return nil }

// Eval implements core.Shader.  Performs all shading for the surface point in sg.  May trace
// rays and shadow rays.
func (sh *ShaderStd) Eval(sg *core.ShaderContext) {

	//fmt.Printf("%v %v %v %v\n", sg.DdDdx, sg.DdNdx, sg.DdDdy, sg.DdNdy)
	/*	deltaTx := m.Vec2Scale(sg.Image.PixelDelta[0], sg.Dduvdx)
		deltaTy := m.Vec2Scale(sg.Image.PixelDelta[1], sg.Dduvdy)

		lod := m.Log2(m.Max(m.Vec2Length(deltaTx), m.Vec2Length(deltaTy)))
		//return
		fmt.Printf("lod: %v\n", lod)

		if lod >= 0 {
			sg.OutRGB[0] = m.Floor(lod) * 10
		} else {
			sg.OutRGB[0] = 100
			sg.OutRGB[1] = 100
		}
		return
	*/
	if sg.Level > 3 {
		return
	}

	// Construct a tangent space
	V := m.Vec3Cross(sg.N, sg.DdPdu)

	if m.Vec3Length2(V) < 0.1 {
		V = m.Vec3Cross(sg.N, sg.DdPdv)
	}
	V = m.Vec3Normalize(V)
	U := m.Vec3Cross(sg.N, V)

	diffRoughness := float32(0.3)

	if sh.DiffuseRoughness != nil {
		diffRoughness = sh.DiffuseRoughness.Float32(sg)
	}

	var _ = diffRoughness
	diffBrdf := bsdf.NewOrenNayar(sg.Lambda, m.Vec3Neg(sg.Rd), diffRoughness, U, V, sg.N)
	//diffBrdf := bsdf.NewLambert(sg.Lambda, m.Vec3Neg(sg.Rd), U, V, sg.N)

	var diffContrib colour.RGB

	var diffColour colour.RGB

	if sh.DiffuseColour != nil {
		diffColour = sh.DiffuseColour.RGB(sg)
	}

	diffWeight := float32(0)
	spec1Weight := float32(0)

	if sh.DiffuseStrength != nil {
		diffWeight = sh.DiffuseStrength.Float32(sg)
	}

	if sh.Spec1Strength != nil {
		spec1Weight = sh.Spec1Strength.Float32(sg)
	}

	totalWeight := diffWeight + spec1Weight
	diffWeight /= totalWeight
	spec1Weight /= totalWeight

	if totalWeight == 0.0 {
		panic(fmt.Sprintf("Shader %v has no weight", sh.Name()))
	}

	if diffWeight > 0.0 {

		sg.LightsPrepare()

		for sg.NextLight() {

			if sg.Lp.DiffuseShadeMult() > 0.0 {

				// In this example the brdf passed is an interface
				// allowing sampling, pdf and bsdf eval
				col := sg.EvaluateLightSamples(diffBrdf)
				col.Mul(diffColour)
				diffContrib.Add(col)
			}

		}

		diffContrib.Scale(diffWeight)
	}

	ior := float32(1.7)

	if sh.IOR != nil {
		ior = sh.IOR.Float32(sg)
	}

	var fresnel core.Fresnel

	switch sh.spec1FresnelModel {
	case fr.DielectricModel:
		fresnel = fr.NewDielectric(ior)
	case fr.ConductorModel:

		refl := colour.RGB{0.5, 0.5, 0.5}
		edge := colour.RGB{0.5, 0.5, 0.5}

		if sh.Spec1FresnelRefl != nil {
			refl = sh.Spec1FresnelRefl.RGB(sg)
		}

		if sh.Spec1FresnelEdge != nil {
			refl = sh.Spec1FresnelEdge.RGB(sg)
		}

		fresnel = fr.NewConductor(0, refl, edge)
	}

	var spec1Contrib colour.RGB

	if spec1Weight > 0.0 {
		spec1Roughness := float32(0.5)

		if sh.Spec1Roughness != nil {
			spec1Roughness = sh.Spec1Roughness.Float32(sg)
		}

		var spec1Colour colour.RGB

		if sh.Spec1Colour != nil {
			spec1Colour = sh.Spec1Colour.RGB(sg)
		}

		var spec1BRDF core.BSDF

		if spec1Roughness == 0.0 {
			spec1BRDF = bsdf.NewSpecular(sg, m.Vec3Neg(sg.Rd), fresnel, U, V, sg.N)
		} else {
			spec1BRDF = bsdf.NewMicrofacetGGX(sg, m.Vec3Neg(sg.Rd), fresnel, spec1Roughness, U, V, sg.N)
		}

		//	Kr := fresnel.Kr(m.Vec3DotAbs(sg.N, m.Vec3Neg(sg.Rd)))
		var samp core.TraceSample
		ray := sg.NewRay()

		spec1Samples := 4

		if spec1Roughness == 0.0 {
			spec1Samples = 1
		}

		for i := 0; i < spec1Samples; i++ {
			idx := uint64(sg.I*spec1Samples /*+ sg.Sample*/ + i)
			r0 := ldseq.VanDerCorput(idx, sg.Scramble[0])
			r1 := ldseq.Sobol(idx, sg.Scramble[1])

			spec1OmegaO := spec1BRDF.Sample(r0, r1)
			pdf := spec1BRDF.PDF(spec1OmegaO)

			if m.Vec3Dot(spec1OmegaO, sg.Ng) <= 0.0 {
				continue

			}
			//fmt.Printf("%v %v\n", spec1Omega, m.Vec3Length(spec1Omega))

			ray.Init(core.RayTypeReflected, sg.OffsetP(1), spec1OmegaO, m.Inf(1), sg.Level+1, sg)

			if core.Trace(ray, &samp) {

				rho := spec1BRDF.Eval(spec1OmegaO)

				rho.Scale(1.0 / float32(pdf))

				col := rho.ToRGB()

				//fmt.Printf("%v %v\n", col, samp.Colour)
				col.Mul(spec1Colour)
				col.Mul(samp.Colour)

				for k := range col {
					if col[k] < 0 || math.IsNaN(float64(col[k])) {
						col[k] = 0
					}
				}
				spec1Contrib.Add(col)
			}

		}

		sg.ReleaseRay(ray)

		spec1Contrib.Scale(spec1Weight / float32(spec1Samples))

		sg.LightsPrepare()

		for sg.NextLight() {

			//			if sg.Lp.DiffuseShadeMult() > 0.0 {

			// In this example the brdf passed is an interface
			// allowing sampling, pdf and bsdf eval
			col := sg.EvaluateLightSamples(spec1BRDF)
			col.Mul(spec1Colour)
			spec1Contrib.Add(col)
			//			}

		}

	}

	contrib := colour.RGB{}

	emissContrib := sh.EvalEmission(sg, m.Vec3Neg(sg.Rd))

	contrib.Add(emissContrib)
	contrib.Add(diffContrib)
	contrib.Add(spec1Contrib)

	sg.OutRGB = contrib
}

// EvalEmission implements core.Shader.
func (sh *ShaderStd) EvalEmission(sg *core.ShaderContext, omegaO m.Vec3) colour.RGB {

	var emissColour colour.RGB
	var emissStrength float32 = 0

	if sh.EmissionColour != nil {
		emissColour = sh.EmissionColour.RGB(sg)
	}

	if sh.EmissionStrength != nil {
		emissStrength = sh.EmissionStrength.Float32(sg)
	} else {
		return colour.RGB{}
	}

	emissColour.Scale(emissStrength)
	return emissColour
}

func init() {
	nodes.Register("ShaderStd", func() (core.Node, error) {

		sh := &ShaderStd{}
		return sh, nil
	})
}
