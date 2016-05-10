// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/jamiec7919/vermeer/colour"
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

type BRDF interface {
	Sample(r0, r1 float64) m.Vec3
	Eval(omega_o m.Vec3) colour.Spectrum
	PDF(omega_o m.Vec3) float64
}

type ShaderGlobals struct {
	X, Y        int     // raster positions
	Sx, Sy      float32 // screen space [-1,1]x[-1,1]
	RayType     uint16  // Ray type
	Depth       uint8
	I, NSamples int     // sample index, total samples
	Weight      float32 // Sample weight
	Lambda      float32 //Wavelength
	Time        float32
	Ro, Rd      m.Vec3         // Ray origin and direction
	Rl          float64        // Ray length (|Ro-P|)
	ElemId      uint32         // Element ID (triangle, curve etc.)
	Prim        Primitive      // primitive pointer
	Psg         *ShaderGlobals // Parent (last shaded)
	Shader      Material

	Po, P, Poffset m.Vec3 // Shading point in object/world space

	N, Nf, Ng, Ngf, Ns m.Vec3 // Shading normal, face-forward shading normal, geometric normal, face-forward geom normal, smoothed normal (N without bump)
	DdPdu, DdPdv       m.Vec3 // Derivative vectors
	Bu, Bv             float32
	U, V               float32 // Surface params

	Lights []Light         // Array of active lights in this shading context
	Lp     Light           // Light pointer (current light)
	Ldist  float32         // distance from P to light source
	Ld     m.Vec3          // Incident direction
	Li     colour.Spectrum // incoming intensity
	Liu    colour.Spectrum // unoccluded incoming

	Area float32

	OutRGB colour.RGB

	rnd *rand.Rand
}

func (sg *ShaderGlobals) Rand() *rand.Rand {
	return sg.rnd
}

// Offset the surface point out from surface by about 1ulp
// Pass -ve value to push point 'into' surface (for transmission)
func (r *ShaderGlobals) OffsetP(dir int) m.Vec3 {
	if dir < 0 {
		r.Poffset = m.Vec3Neg(r.Poffset)

	}
	po := m.Vec3Add(r.P, r.Poffset)

	// round po away from p
	for i := range po {
		//log.Printf("%v %v %v", i, offset[i], po[i])
		if r.Poffset[i] > 0 {
			po[i] = m.NextFloatUp(po[i])
		} else if r.Poffset[i] < 0 {
			po[i] = m.NextFloatDown(po[i])
		}
		//log.Printf("%v %v", i, po[i])
	}

	return po
}

// The texture derivatives Pu & Pv should already be computed.
func (sg *ShaderGlobals) SetupTangentSpace_(Ns m.Vec3) {
	sg.N = m.Vec3Normalize(Ns)
	sg.DdPdu = m.Vec3Normalize(m.Vec3Cross(sg.N, sg.DdPdu))
	sg.DdPdv = m.Vec3Normalize(m.Vec3Cross(sg.N, sg.DdPdu))

}

func (sg *ShaderGlobals) WorldToTangent(v m.Vec3) m.Vec3 {
	V := m.Vec3Normalize(m.Vec3Cross(sg.N, sg.DdPdu))
	U := m.Vec3Cross(sg.N, V)
	return m.Vec3BasisProject(U, V, sg.N, v)
}

func (sg *ShaderGlobals) TangentToWorld(v m.Vec3) m.Vec3 {
	V := m.Vec3Normalize(m.Vec3Cross(sg.N, sg.DdPdu))
	U := m.Vec3Cross(sg.N, V)
	return m.Vec3BasisExpand(U, V, sg.N, v)
}

func (sg *ShaderGlobals) ViewDirection() m.Vec3 {
	return sg.WorldToTangent(m.Vec3Neg(sg.Rd))
}

func (sg *ShaderGlobals) LightsPrepare() {
	sg.I = 0

	sg.Lights = make([]Light, 0, len(grc.scene.lights))

	for _, light := range grc.scene.lights {
		sg.Lights = append(sg.Lights, light)
	}

}

func GetMaterial(id int32) Material {
	return grc.GetMaterial(id)
}

func (sg *ShaderGlobals) LightsGetSample() bool {

retry:
	if sg.I < len(sg.Lights) {
		sg.Lp = sg.Lights[sg.I]
		sg.I++

		if sg.Lp.SampleArea(sg) == nil {
		} else {
			goto retry
		}

		return true
	} else {
		return false
	}
}

var shadowRays int

func (sg *ShaderGlobals) EvaluateLightSample(brdf BRDF) colour.RGB {
	// The brdf returns directions in the tangent space
	ray := new(RayData)

	ray.Init(RAY_SHADOW, sg.OffsetP(1), m.Vec3Scale(sg.Ldist*(1.0-VisRayEpsilon), sg.Ld), 1.0, sg)

	shadowRays++
	if !TraceProbe(ray, &ShaderGlobals{}) { // for shadow rays sg is not modified so to avoid allocations reuse it here

		rho := brdf.Eval(sg.WorldToTangent(sg.Ld))

		rho.Mul(sg.Liu)
		rho.Scale(sg.Weight)

		r, g, b := rho.ToRGB()
		return colour.RGB{r, g, b}
	} else {
		return colour.RGB{}
	}

}

func (sg *ShaderGlobals) GlossySample(brdf BRDF) m.Vec3 {
	omega_o := brdf.Sample(sg.rnd.Float64(), sg.rnd.Float64())

	// Eval should take the PDF into account.. ??
	sg.Weight = 1.0 // float32(brdf.PDF(omega_o))
	return omega_o
}
