// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/jamiec7919/vermeer/colour"
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

// BSDF represents a BRDF or BTDF that can be sampled.  The core API only considers the incoming direction
// so all other parameters should be stored in the concrete instance of the individual BRDFs.
// This is used for Multiple Importance Sampling with light sources.
type BSDF interface {
	// Sample returns a direction given two (quasi)random numbers.
	Sample(r0, r1 float64) m.Vec3
	// Eval evaluates the BSDF given the incoming direction. Returns value multiplied by cosine
	// of angle between omegaO and the normal.
	Eval(omegaO m.Vec3) colour.Spectrum
	// PDF returns the probability density function for the given sample.
	PDF(omegaO m.Vec3) float64
}

// Fresnel represents a Fresnel model.
type Fresnel interface {
	// Kr returns the fresnel value.  cos_theta is the clamped dot product of
	// view direction and surface normal.
	Kr(cos_theta float32) colour.RGB
}

// ShaderGlobals encapsulates all of the data needed for evaluating shaders.
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

// Rand returns the rng in use.
//
// Deprecated: although still in use, should be removed once new sampling is introduced.
func (sg *ShaderGlobals) Rand() *rand.Rand {
	return sg.rnd
}

// OffsetP returns the intersection point pushed out from surface by about 1 ulp.
// Pass -ve value to push point 'into' surface (for transmission).
func (sg *ShaderGlobals) OffsetP(dir int) m.Vec3 {

	var pofs m.Vec3
	if dir < 0 {
		pofs = m.Vec3Neg(sg.Poffset)

	} else {
		pofs = sg.Poffset
	}

	po := m.Vec3Add(sg.P, pofs)

	// round po away from p
	for i := range po {
		//log.Printf("%v %v %v", i, offset[i], po[i])
		if pofs[i] > 0 {
			po[i] = m.NextFloatUp(po[i])
		} else if pofs[i] < 0 {
			po[i] = m.NextFloatDown(po[i])
		}
		//log.Printf("%v %v", i, po[i])
	}

	return po
}

// WorldToTangent projects the direction v into the tangent space
// formed from the shading normal N and texture derivative tangents.
func (sg *ShaderGlobals) WorldToTangent(v m.Vec3) m.Vec3 {
	V := m.Vec3Cross(sg.N, sg.DdPdu)

	if m.Vec3Length2(V) < 0.1 {
		V = m.Vec3Cross(sg.N, sg.DdPdv)
	}
	V = m.Vec3Normalize(V)
	U := m.Vec3Cross(sg.N, V)
	return m.Vec3BasisProject(U, V, sg.N, v)
}

// TangentToWorld projects the direction v into world space
// based on the tangent space formed from the shading normal N and texture derivative tangents.
func (sg *ShaderGlobals) TangentToWorld(v m.Vec3) m.Vec3 {
	V := m.Vec3Cross(sg.N, sg.DdPdu)

	if m.Vec3Length2(V) < 0.1 {
		V = m.Vec3Cross(sg.N, sg.DdPdv)
	}
	V = m.Vec3Normalize(V)
	U := m.Vec3Cross(sg.N, V)
	return m.Vec3BasisExpand(U, V, sg.N, v)
}

// ViewDirection returns the view direction (-ve ray direction projected into tangent space).
func (sg *ShaderGlobals) ViewDirection() m.Vec3 {
	return m.Vec3Normalize(sg.WorldToTangent(m.Vec3Neg(sg.Rd)))
}

// LightsPrepare initialises the lighting loop.
func (sg *ShaderGlobals) LightsPrepare() {
	sg.I = 0

	sg.Lights = make([]Light, 0, len(grc.scene.lights))

	for _, light := range grc.scene.lights {
		sg.Lights = append(sg.Lights, light)
	}

}

// GetMaterial returns the shader for the given id.
func GetMaterial(id int32) Material {
	return grc.GetMaterial(id)
}

// LightsGetSample should be called in a loop and will setup the globals for the next light
// sample and return true.  False will be returned when no more samples are available.
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
	}

	return false

}

// EvaluateLightSample will evaluate the MIS sample for the current light sample and given BRDF.
func (sg *ShaderGlobals) EvaluateLightSample(brdf BSDF) colour.RGB {
	// The brdf returns directions in the tangent space
	ray := new(RayData)

	if m.Vec3Dot(sg.Ld, sg.Ng) < 0 {
		ray.Init(RAY_SHADOW, sg.OffsetP(-1), m.Vec3Scale(sg.Ldist*(1.0-VisRayEpsilon), sg.Ld), 1.0, sg)
	} else {
		ray.Init(RAY_SHADOW, sg.OffsetP(1), m.Vec3Scale(sg.Ldist*(1.0-VisRayEpsilon), sg.Ld), 1.0, sg)

	}

	if !TraceProbe(ray, &ShaderGlobals{}) { // for shadow rays sg is not modified so to avoid allocations reuse it here

		rho := brdf.Eval(sg.WorldToTangent(sg.Ld))

		rho.Mul(sg.Liu)
		rho.Scale(sg.Weight)

		r, g, b := rho.ToRGB()
		return colour.RGB{r, g, b}
	}

	return colour.RGB{}

}

// GlossySample generates a sample from the BSDF and sets the globals weight.
func (sg *ShaderGlobals) GlossySample(brdf BSDF) m.Vec3 {
	omegaO := brdf.Sample(sg.rnd.Float64(), sg.rnd.Float64())

	// Eval should take the PDF into account.. ??
	sg.Weight = 1.0 / float32(brdf.PDF(omegaO))
	return omegaO
}
