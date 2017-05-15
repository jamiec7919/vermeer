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
	//"math"
)

// ShaderStd is the default surface shader.
type ShaderStd struct {
	NodeDef core.NodeDef `node:"-"`
	MtlName string       `node:"Name"`

	EmissionColour   param.RGBUniform     `node:",opt"`
	EmissionStrength param.Float32Uniform `node:",opt"`

	OneSided bool `node:",opt"` // One or two sided (defaults to 2 which means has different properties depending on which side ray approaches)

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

	IsTransmissive bool `node:"Transmissive,opt"`
	Priority       int  `node:",opt"` // Must be <= 255 (8 bits), lower is higher priority
	interiorID     uint64
	TransColour    param.RGBUniform     `node:",opt"` // Colour parameter
	TransStrength  param.Float32Uniform `node:",opt"` // Weight parameter
	IOR            param.Float32Uniform `node:",opt"`
}

func refract2(omegaR, N m.Vec3, ior float32) m.Vec3 {
	eta := 1.0 / ior

	c := m.Vec3Dot(omegaR, N)

	a := eta*c - m.Sqrt(1.0+eta*(c*c-1.0))

	return m.Vec3Sub(m.Vec3Scale(a, N), m.Vec3Scale(eta, omegaR))

}

func reflect(omegaR, N m.Vec3) (omegaO m.Vec3) {
	//omegaO = m.Vec3Add(m.Vec3Neg(omegaR), m.Vec3Scale(2*m.Vec3DotAbs(omegaR, N), N))

	omegaO = m.Vec3Add(omegaR, m.Vec3Scale(2.0*m.Vec3Dot(N, omegaR), N))

	return
}

func refract(d, N m.Vec3, ior float32) (bool, m.Vec3) {
	// ior = n/n_t or inverse
	dotN := m.Vec3Dot(d, N)

	sq := 1 - ior*ior*(1-dotN*dotN)

	// Total internal reflection
	if sq < 0 {
		//fmt.Printf("sq: %v\n", sq)
		return false, m.Vec3{}
	}

	omega := m.Vec3Sub(m.Vec3Scale(ior, m.Vec3Sub(d, m.Vec3Scale(dotN, N))), m.Vec3Scale(m.Sqrt(sq), N))
	//fmt.Printf("%v %v %v\n", d, omega, N)
	return true, omega
}

var interiorID uint64 = 1

// NextInteriorID returns a unique interior ID
// TODO: add locking
func NextInteriorID() uint64 {
	id := interiorID
	interiorID++
	return id << 8
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

	if sh.IsTransmissive {
		sh.interiorID = NextInteriorID() | uint64(sh.Priority)
	}

	return nil
}

// PostRender is a core.Node method.
func (sh *ShaderStd) PostRender() error { return nil }

// Eval implements core.Shader.  Performs all shading for the surface point in sg.  May trace
// rays and shadow rays.
func (sh *ShaderStd) Eval(sg *core.ShaderContext) bool {

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
	if sg.Level > 8 {
		return true
	}

	ior := float32(1.7)

	if sh.IOR != nil {
		ior = sh.IOR.Float32(sg)
	}

	transEnter := true     // Assume we're entering surface
	var ior1, ior2 float32 // ior1 is the medium we're going from, ior2 is medium we're entering

	if sh.IsTransmissive {
		if m.Vec3Dot(sg.Rd, sg.Ng) > 0 {
			// Nope we're exiting
			transEnter = false

			ior1 = ior
			ior2 = 1.00029

			// We're leaving the surface, ior2 should be set to the highest priority surface in
			// the InteriorList.  If we are not the highest priority then remove and false hit.
			// Otherwise, don't add us as need to do that after checking for TIR.
			highestPriority := uint64(255)

			for i := range sg.InteriorList {
				if sg.InteriorList[i].InteriorID != 0 && sg.InteriorList[i].InteriorID != sh.interiorID {
					priority := sg.InteriorList[i].InteriorID & 0xff

					if priority < highestPriority {
						highestPriority = priority
						ior2 = sg.InteriorList[i].IOR
					}
				}
			}

			if uint64(sh.Priority) > highestPriority {
				// We aren' highest priority so remove and false hit
				for i := range sg.InteriorList {
					if sg.InteriorList[i].InteriorID == sh.interiorID {
						sg.InteriorList[i].InteriorID = 0
					}
				}
				return false
			}
		} else {
			ior1 = 1.00029
			ior2 = ior

			// We're entering the surface, ior1 should be set to the highest priority surface in
			// the InteriorList.  If we are not the highest priority then add and false hit.

			highestPriority := uint64(255)

			for i := range sg.InteriorList {
				if sg.InteriorList[i].InteriorID != 0 {
					priority := sg.InteriorList[i].InteriorID & 0xff

					if priority < highestPriority {
						highestPriority = priority
						ior1 = sg.InteriorList[i].IOR
					}
				}
			}

			if uint64(sh.Priority) > highestPriority {
				// We aren't highest priority so add and false hit
				sg.InteriorList = append(sg.InteriorList, core.InteriorListEntry{IOR: ior, InteriorID: sh.interiorID})

				return false
			}

		}
	}

	// Construct a tangent space
	V := m.Vec3Cross(sg.N, sg.DdPdu)

	if m.Vec3Length2(V) < 0.1 {
		V = m.Vec3Cross(sg.N, sg.DdPdv)
	}
	V = m.Vec3Normalize(V)
	U := m.Vec3Normalize(m.Vec3Cross(sg.N, V))

	diffRoughness := float32(0.5)

	if sh.DiffuseRoughness != nil {
		diffRoughness = sh.DiffuseRoughness.Float32(sg)
	}

	var _ = diffRoughness
	diffBrdf := bsdf.NewOrenNayar(sg.Lambda, m.Vec3Neg(sg.Rd), diffRoughness, U, V, sg.N)
	//diffBrdf := bsdf.NewLambert(sg.Lambda, m.Vec3Neg(sg.Rd), U, V, sg.N)

	var diffContrib colour.Spectrum
	diffContrib.Lambda = sg.Lambda

	var diffColour colour.Spectrum
	diffColour.Lambda = sg.Lambda

	if sh.DiffuseColour != nil {
		diffColour.FromRGB(sh.DiffuseColour.RGB(sg))
	}

	var diffWeight, spec1Weight, transWeight, totalWeight float32

	if sh.DiffuseStrength != nil {
		diffWeight = sh.DiffuseStrength.Float32(sg)
	}

	if sh.Spec1Strength != nil {
		spec1Weight = sh.Spec1Strength.Float32(sg)
	}

	if sh.TransStrength != nil {
		transWeight = sh.TransStrength.Float32(sg)
	}

	if !sh.IsTransmissive {
		totalWeight = diffWeight + spec1Weight
		diffWeight /= totalWeight
		spec1Weight /= totalWeight
		transWeight = 0
	} else {
		totalWeight = transWeight + spec1Weight
		transWeight /= totalWeight
		spec1Weight /= totalWeight
		diffWeight = 0

	}

	if totalWeight == 0.0 {
		panic(fmt.Sprintf("Shader %v has no weight", sh.Name()))
	}

	if diffWeight > 0.0 && !sh.IsTransmissive {

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

	var spec1Contrib colour.Spectrum
	spec1Contrib.Lambda = sg.Lambda

	if spec1Weight > 0.0 {
		spec1Roughness := float32(0.5)

		if sh.Spec1Roughness != nil {
			spec1Roughness = sh.Spec1Roughness.Float32(sg)
		}

		var spec1Colour colour.Spectrum
		spec1Colour.Lambda = sg.Lambda

		if sh.Spec1Colour != nil {
			spec1Colour.FromRGB(sh.Spec1Colour.RGB(sg))
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

		spec1Samples := 0

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

				//col := rho.ToRGB()

				//fmt.Printf("%v %v\n", col, samp.Colour)
				rho.Mul(spec1Colour)
				rho.Mul(samp.Spectrum)

				//for k := range col {
				//	if col[k] < 0 || math.IsNaN(float64(col[k])) {
				//		col[k] = 0
				//	}
				//}
				spec1Contrib.Add(rho)
			}

		}

		sg.ReleaseRay(ray)

		if spec1Samples > 0 {
			spec1Contrib.Scale(1 / float32(spec1Samples))
		}

		if spec1Roughness > 0.0 { // No point doing direct lighting for mirror surfaces!
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

		spec1Contrib.Scale(spec1Weight)
	}

	var transContrib colour.Spectrum
	transContrib.Lambda = sg.Lambda

	if sh.IsTransmissive {
		ray := sg.NewRay()
		var samp core.TraceSample

		var transColour colour.Spectrum
		transColour.Lambda = sg.Lambda

		if sh.TransColour != nil {
			transColour.FromRGB(sh.TransColour.RGB(sg))
		}

		if transEnter {
			ok, omega := refract(sg.Rd, sg.N, ior1/ior2)

			if ok {
				// We've entered the surface so add to the interior list

				ray.Init(core.RayTypeRefracted, sg.OffsetP(-1), omega, m.Inf(1), sg.Level+1, sg)
				ray.InteriorList = append(ray.InteriorList, core.InteriorListEntry{IOR: ior, InteriorID: sh.interiorID})

				if core.Trace(ray, &samp) {
					transContrib = samp.Spectrum
					transContrib.Mul(transColour)
					//					transContrib = samp.Colour
					//					absorb := colour.RGB{m.Exp(-(1 - transColour[0]) * ray.Tclosest),
					//						m.Exp(-(1 - transColour[1]) * ray.Tclosest),
					//						m.Exp(-(1 - transColour[2]) * ray.Tclosest),
					//					}
					//					transContrib.Mul(absorb)
				}

			}

		} else {
			ok, omega := refract(sg.Rd, m.Vec3Neg(sg.N), ior1/ior2)

			if ok {
				// we've left the surface so remove the interior

				ray.Init(core.RayTypeRefracted, sg.OffsetP(1), omega, m.Inf(1), sg.Level+1, sg)

				for i := range ray.InteriorList {
					if ray.InteriorList[i].InteriorID == sh.interiorID {
						ray.InteriorList[i].InteriorID = 0
					}
				}

				if core.Trace(ray, &samp) {
					transContrib = samp.Spectrum
					//transContrib.Mul(transColour)
				}
			} else {
				omega := reflect(sg.Rd, m.Vec3Neg(sg.N))

				ray.Init(core.RayTypeRefracted, sg.OffsetP(-1), omega, m.Inf(1), sg.Level+1, sg)

				if core.Trace(ray, &samp) {
					transContrib = samp.Spectrum
					transWeight = spec1Weight // This is really a specular bounce
				}
			}
		}

		sg.ReleaseRay(ray)
		transContrib.Scale(transWeight)

	}

	contrib := colour.Spectrum{Lambda: sg.Lambda}

	emissContrib := sh.EvalEmission(sg, m.Vec3Neg(sg.Rd))

	contrib.Add(emissContrib)
	contrib.Add(diffContrib)
	contrib.Add(spec1Contrib)
	contrib.Add(transContrib)

	//sg.OutRGB = contrib.ToRGB()
	sg.Out = contrib

	return true
}

// EvalEmission implements core.Shader.
func (sh *ShaderStd) EvalEmission(sg *core.ShaderContext, omegaO m.Vec3) colour.Spectrum {

	emissColour := colour.Spectrum{Lambda: sg.Lambda}
	var emissStrength float32 = 0

	if sh.EmissionColour != nil {
		emissColour.FromRGB(sh.EmissionColour.RGB(sg))
	}

	if sh.EmissionStrength != nil {
		emissStrength = sh.EmissionStrength.Float32(sg)
	} else {
		return colour.Spectrum{Lambda: sg.Lambda}
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
