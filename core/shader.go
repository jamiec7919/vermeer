// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"github.com/jamiec7919/vermeer/colour"
	m "github.com/jamiec7919/vermeer/math"
)

// BSDF represents a BRDF or BTDF that can be sampled.  The core API only considers the incoming direction
// so all other parameters should be stored in the concrete instance of the individual BRDFs.
// This is used for Multiple Importance Sampling with light sources.
type BSDF interface {
	// Sample returns a direction given two (quasi)random numbers.
	Sample(r0, r1 float64) m.Vec3
	// Eval evaluates the BSDF given the incoming direction. Returns value multiplied by cosine
	// of angle between omegaO and the normal. NOTE: omegaO is in WORLD space.
	Eval(omegaO m.Vec3) colour.Spectrum
	// PDF returns the probability density function for the given sample. NOTE: omegaO is in WORLD space.
	PDF(omegaO m.Vec3) float64
}

// Fresnel represents a Fresnel model.
type Fresnel interface {
	// Kr returns the fresnel value.  cos_theta is the clamped dot product of
	// view direction and surface normal.
	Kr(cosTheta float32) colour.RGB
}

// Shader represents a surface shader (Note: this will be renamed to Shader or SurfaceShader).
type Shader interface {
	// Eval evaluates the shader and returns values in sh.OutXXX members.
	Eval(sc *ShaderContext)

	// EvalEmission evaluates the shader and returns emission value.
	EvalEmission(sc *ShaderContext, omegaO m.Vec3) colour.RGB
}

// ShaderContext encapsulates all of the data needed for evaluating shaders.
// These should only ever be created with NewShaderContext.
type ShaderContext struct {
	X, Y                            int32   // raster positions
	Sx, Sy                          float32 // screen space [-1,1]x[-1,1]
	RayType                         uint16  // Ray type
	Level                           uint8   // Recursive level of sample (0 for first hit)
	I, Sample, NSamples             int     // pixel sample index, sample index, total samples
	SampleScramble, SampleScramble2 uint64
	Weight                          float32 // Sample weight
	Lambda                          float32 // Wavelength
	Time                            float32
	Ro, Rd                          m.Vec3         // Ray origin and direction
	Rl                              float64        // Ray length (|Ro-P|)
	ElemID                          uint32         // Element ID (triangle, curve etc.)
	Geom                            Geom           // Geom pointer
	Psc                             *ShaderContext // Parent (last shaded)
	Shader                          Shader

	Transform, InvTransform m.Matrix4

	Po, P, Poffset m.Vec3 // Shading point in object/world space

	N, Nf, Ng, Ngf, Ns m.Vec3  // Shading normal, face-forward shading normal, geometric normal, face-forward geom normal, smoothed normal (N without bump)
	DdPdu, DdPdv       m.Vec3  // Derivative vectors
	Bu, Bv             float32 // Barycentric U&V
	U, V               float32 // Surface params

	// Ray differentials
	DdPdx, DdPdy m.Vec3 // Ray differential
	DdDdx, DdDdy m.Vec3 // Ray differential
	DdNdx, DdNdy m.Vec3 // Normal derivatives
	Ddudx, Ddudy m.Vec3 // texture/surface param derivs
	Ddvdx, Ddvdy m.Vec3 // texture/surface param derivs

	Lights   []Light // Array of active lights in this shading context
	Lidx     int     // Index of current light
	Lsamples []LightSample
	Lp       Light           // Light pointer (current light)
	Ldist    float32         // distance from P to light source
	Ld       m.Vec3          // Incident direction
	Li       colour.Spectrum // incoming intensity
	Liu      colour.Spectrum // unoccluded incoming

	Area float32

	OutRGB colour.RGB

	task *RenderTask
	next *ShaderContext // Pool link
	priv *shaderPrivate
}

type shaderPrivate struct {
	next *shaderPrivate // Pool link
}

// NewShaderContext returns a new context from the pool.  Should not create these manually.
func (sc *ShaderContext) NewShaderContext() *ShaderContext {
	return sc.task.NewShaderContext()
}

// ReleaseShaderContext returns a context to the pool.
func (sc *ShaderContext) ReleaseShaderContext(sc2 *ShaderContext) {
	sc.task.ReleaseShaderContext(sc2)
}

// NewRay returns a new ray from the pool.
func (sc *ShaderContext) NewRay() *Ray {
	return sc.task.NewRay()
}

// ReleaseRay returns ray to pool
func (sc *ShaderContext) ReleaseRay(ray *Ray) {
	sc.task.ReleaseRay(ray)
}

// OffsetP returns the intersection point pushed out from surface by about 1 ulp.
// Pass -ve value to push point 'into' surface (for transmission).
func (sc *ShaderContext) OffsetP(dir int) m.Vec3 {

	var pofs m.Vec3
	if dir < 0 {
		pofs = m.Vec3Neg(sc.Poffset)

	} else {
		pofs = sc.Poffset
	}

	po := m.Vec3Add(sc.P, pofs)

	// round po away from p
	for i := range po {
		//log.Printf("%v %v %v", i, offset[i], po[i])
		if pofs[i] > 0 {
			po[i] = m.NextFloatUp(po[i])
		} else if pofs[i] < -0 {
			po[i] = m.NextFloatDown(po[i])
		}
		//log.Printf("%v %v", i, po[i])
	}

	return po
}

// LightsPrepare initialises the lighting loop.
func (sc *ShaderContext) LightsPrepare() {
	sc.Sample = 0
	sc.SampleScramble = uint64(sc.task.rand.Int63())  // This no longer needs to be in sc
	sc.SampleScramble2 = uint64(sc.task.rand.Int63()) // as the lights generate the n samples in SampleArea.
	//sc.SampleScramble = 0
	scene.LightsPrepare(sc)

	sc.Lidx = -1 // Must be -1 as it is updated first thing in LightsGetSample.

	if sc.Lsamples != nil {
		sc.Lsamples = sc.Lsamples[:0]
	}
	// Should take light.NumSamples samples from each light
}

// LightsGetSample should be called in a loop and will setup the context for the next light
// sample and return true.  False will be returned when no more samples are available.
func (sc *ShaderContext) LightsGetSample() bool {

	if sc.Sample >= len(sc.Lsamples) {
		sc.Lidx++

		if sc.Lidx >= len(sc.Lights) { // All done
			return false
		}

		n := sc.Lights[sc.Lidx].NumSamples(sc)

		if cap(sc.Lsamples) < n {
			sc.Lsamples = make([]LightSample, 0, n)
		} else {
			sc.Lsamples = sc.Lsamples[:0]
		}

		sc.Sample = 0
		sc.SampleScramble = uint64(sc.task.rand.Int63())
		sc.SampleScramble2 = uint64(sc.task.rand.Int63())
		sc.Lp = sc.Lights[sc.Lidx]
		sc.Lp.SampleArea(sc, n)

	}

	if sc.Sample >= len(sc.Lsamples) {
		// This will only happen if there is a problem in sc.Lp.SampleArea. Recursing
		// will hit the test at the top and break the recursion if we run out of lights.
		return sc.LightsGetSample()
	}

	sc.Liu = sc.Lsamples[sc.Sample].Liu
	sc.Ld = sc.Lsamples[sc.Sample].Ld
	sc.Ldist = sc.Lsamples[sc.Sample].Ldist
	sc.Weight = float32(sc.Lsamples[sc.Sample].Weight)
	sc.Sample++
	return true
}

// EvaluateLightSample will evaluate the MIS sample for the current light sample and given BRDF.
func (sc *ShaderContext) EvaluateLightSample(brdf BSDF) colour.RGB {
	// The brdf returns directions in the tangent space
	ray := sc.NewRay()
	chsc := sc.NewShaderContext()
	if false {
		fmt.Printf("%v %v\n", sc.P, sc.OffsetP(1))
	}
	if m.Vec3Dot(sc.Ld, sc.Ng) < 0 {
		ray.Init(RayTypeShadow, sc.OffsetP(-1), m.Vec3Scale(sc.Ldist*(1.0-ShadowRayEpsilon), sc.Ld), 1.0, 0, sc.Lambda, sc.Time)
	} else {
		ray.Init(RayTypeShadow, sc.OffsetP(1), m.Vec3Scale(sc.Ldist*(1.0-ShadowRayEpsilon), sc.Ld), 1.0, 0, sc.Lambda, sc.Time)

	}

	if !TraceProbe(ray, chsc) {

		rho := brdf.Eval(sc.Ld)

		//fmt.Printf("%v %v %v : \n", rho, sc.Liu, sc.Weight)

		rho.Mul(sc.Liu)
		rho.Scale(sc.Weight / float32(len(sc.Lsamples)))

		//fmt.Printf("%v\n\n", rho)
		rgb := rho.ToRGB()

		sc.ReleaseRay(ray)
		sc.ReleaseShaderContext(chsc)
		return rgb
	}

	sc.ReleaseShaderContext(chsc)
	sc.ReleaseRay(ray)
	return colour.RGB{}

}
