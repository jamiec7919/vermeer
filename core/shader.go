// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/jamiec7919/vermeer/colour"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/math/ldseq"
)

// BSDF represents a BRDF or BTDF that can be sampled.  The core API only considers the incoming direction
// so all other parameters should be stored in the concrete instance of the individual BRDFs.
// This is used for Multiple Importance Sampling with light sources.
type BSDF interface {
	// Sample returns a direction given two (quasi)random numbers. NOTE: returned vector in WORLD space.
	Sample(r0, r1 float64) m.Vec3
	// Eval evaluates the BSDF given the incoming direction. Returns value multiplied by cosine
	// of angle between omegaO and the normal. NOTE: omegaO is in WORLD space.
	Eval(omegaO m.Vec3) colour.Spectrum
	// PDF returns the probability density function for the given sample. NOTE: omegaO is in WORLD space.
	PDF(omegaO m.Vec3) float64
}

type BSDFSample struct {
	D      m.Vec3
	Pdf    float64
	Weight float32
	Liu    colour.Spectrum
	Ld     m.Vec3
	Ldist  float32
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
	X, Y                int32     // raster positions
	Sx, Sy              float32   // screen space [-1,1]x[-1,1]
	RayType             uint16    // Ray type
	Level               uint8     // Recursive level of sample (0 for first hit)
	I, Sample, NSamples int       // pixel sample index, sample index, total samples
	Scramble            [2]uint64 // Per pixel scramble values for lights and glossy
	Weight              float32   // Sample weight
	Lambda              float32   // Wavelength
	Time                float32
	Ro, Rd              m.Vec3         // Ray origin and direction
	Rl                  float64        // Ray length (|Ro-P|)
	ElemID              uint32         // Element ID (triangle, curve etc.)
	Geom                Geom           // Geom pointer
	Psc                 *ShaderContext // Parent (last shaded)
	Shader              Shader

	Transform, InvTransform m.Matrix4

	Po, P, Poffset m.Vec3 // Shading point in object/world space

	N, Nf, Ng, Ngf, Ns m.Vec3  // Shading normal, face-forward shading normal, geometric normal, face-forward geom normal, smoothed normal (N without bump)
	DdPdu, DdPdv       m.Vec3  // Derivative vectors
	Bu, Bv             float32 // Barycentric U&V
	U, V               float32 // Surface params

	// Ray differentials
	DdPdx, DdPdy   m.Vec3 // Ray differential
	DdDdx, DdDdy   m.Vec3 // Ray differential
	DdNdx, DdNdy   m.Vec3 // Normal derivatives
	Dduvdx, Dduvdy m.Vec2 // texture/surface param derivs

	Lights   []Light // Array of active lights in this shading context
	Lidx     int     // Index of current light
	Lsamples []LightSample
	Lp       Light // Light pointer (current light)

	Area float32

	Image *Image // Image constant values stored here

	OutRGB      colour.RGB
	OutSpectrum colour.Spectrum

	task *RenderTask
	next *ShaderContext // Pool link
	priv *shaderPrivate
}

type shaderPrivate struct {
	next *shaderPrivate // Pool link
}

// NewShaderContext returns a new context from the pool.  Should not create these manually.
func (sc *ShaderContext) NewShaderContext() *ShaderContext {
	s := sc.task.NewShaderContext()
	s.Image = sc.Image
	return s
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

// ApplyTransform applies Transfomrm to appropriate fields to go from object space into world space.
func (sc *ShaderContext) ApplyTransform() {
	sc.P = m.Matrix4MulPoint(sc.Transform, sc.Po)
	sc.N = m.Vec3Normalize(m.Matrix4MulVec(m.Matrix4Transpose(sc.InvTransform), sc.N))
	sc.Ng = m.Vec3Normalize(m.Matrix4MulVec(m.Matrix4Transpose(sc.InvTransform), sc.Ng))
	sc.DdPdu = m.Vec3Normalize(m.Matrix4MulVec(sc.Transform, sc.DdPdu))
	sc.DdPdv = m.Vec3Normalize(m.Matrix4MulVec(sc.Transform, sc.DdPdv))
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
	scene.LightsPrepare(sc)

	sc.Lidx = -1 // Must be -1 as it is updated first thing in LightsGetSample.

	if sc.Lsamples != nil {
		sc.Lsamples = sc.Lsamples[:0]
	}
}

// NextLight sets up the ShaderContext for the next relevant light and returns true.  If there are
// no further lights will return false.
func (sc *ShaderContext) NextLight() bool {
	sc.Lidx++

	if sc.Lidx < len(sc.Lights) {
		sc.Lp = sc.Lights[sc.Lidx]
		sc.Sample = 0
		// Should take light.NumSamples samples from each light
		sc.NSamples = sc.Lp.NumSamples(sc)

		// Unless we're after first bounce
		if sc.Level > 0 {
			sc.NSamples = 1
		}

		return true
	}

	return false
}

// EvaluateLightSamples will evaluate direct lighting for the current light using MIS and
// return total contribution.  This can be weighted by albedo (colour).
// Will do MIS for diffuse too but just discard any that miss light. Can do BRDF first up to NSamples/2
// then any left over samples will be given to light sampling.
func (sc *ShaderContext) EvaluateLightSamples(bsdf BSDF) colour.RGB {
	var bsdfSamples []BSDFSample
	var col colour.RGB

	if sc.NSamples > 1 {

		for i := 0; i < sc.NSamples/2; i++ {
			idx := uint64(sc.I*(sc.NSamples/2) + sc.Sample + i)
			r0 := ldseq.VanDerCorput(idx, sc.Scramble[0])
			r1 := ldseq.Sobol(idx, sc.Scramble[1])

			omegaO := bsdf.Sample(r0, r1)

			s := BSDFSample{D: omegaO, Pdf: bsdf.PDF(omegaO)}

			if sc.Lp.ValidSample(sc, &s) {

				bsdfSamples = append(bsdfSamples, s)
			}
		}

		if cap(sc.Lsamples) < sc.NSamples/2 {
			sc.Lsamples = make([]LightSample, 0, sc.NSamples/2)
		} else {
			sc.Lsamples = sc.Lsamples[:0]
		}

		sc.Lp.SampleArea(sc, sc.NSamples/2)

		for i := 0; i < sc.NSamples/2; i++ {

		}

		nBSDFSamples := sc.NSamples / 2
		nLightSamples := sc.NSamples / 2

		if len(sc.Lsamples) == 0 { // NOTE this needs thinking about, if we take any samples at all then should count all of them, otherwise use BSDF.
			nLightSamples = 0
		}

		totalSamples := nBSDFSamples + nLightSamples

		if totalSamples == 0 {
			return colour.RGB{}
		}

		//totalSamples = sc.NSamples

		ray := sc.NewRay()
		chsc := sc.NewShaderContext()

		for _, ls := range sc.Lsamples {

			if m.Vec3Dot(ls.Ld, sc.Ng) < 0 {
				ray.Init(RayTypeShadow, sc.OffsetP(-1), m.Vec3Scale(ls.Ldist*(1.0-ShadowRayEpsilon), ls.Ld), 1.0, 0, sc)
			} else {
				ray.Init(RayTypeShadow, sc.OffsetP(1), m.Vec3Scale(ls.Ldist*(1.0-ShadowRayEpsilon), ls.Ld), 1.0, 0, sc)

			}

			if !TraceProbe(ray, chsc) {

				rho := bsdf.Eval(ls.Ld)

				//fmt.Printf("%v %v %v : \n", rho, sc.Liu, sc.Weight)

				rho.Mul(ls.Liu)

				p_hat := float32(nBSDFSamples) * float32(bsdf.PDF(ls.Ld)) / float32(totalSamples)

				p_hat += float32(nLightSamples) * ls.Weight / float32(totalSamples)

				rho.Scale(1.0 / p_hat)

				//fmt.Printf("%v\n\n", rho)
				rgb := rho.ToRGB()

				for k := range rgb {
					if rgb[k] < 0 {
						rgb[k] = 0
					}
				}

				col.Add(rgb)

			}
		}

		for _, bs := range bsdfSamples {

			if m.Vec3Dot(bs.Ld, sc.Ng) < 0 {
				ray.Init(RayTypeShadow, sc.OffsetP(-1), m.Vec3Scale(bs.Ldist*(1.0-ShadowRayEpsilon), bs.Ld), 1.0, 0, sc)
			} else {
				ray.Init(RayTypeShadow, sc.OffsetP(1), m.Vec3Scale(bs.Ldist*(1.0-ShadowRayEpsilon), bs.Ld), 1.0, 0, sc)

			}

			if !TraceProbe(ray, chsc) {

				rho := bsdf.Eval(bs.Ld)

				rho.Mul(bs.Liu)

				p_hat := float32(nBSDFSamples) * float32(bs.Pdf) / float32(totalSamples)

				p_hat += float32(nLightSamples) * bs.Weight / float32(totalSamples)

				rho.Scale(1.0 / p_hat)

				rgb := rho.ToRGB()

				//fmt.Printf("%v %v %v %v %v %v %v\n", sc.X, sc.Y, totalSamples, bs.Pdf, p_hat, rho, rgb)

				for k := range rgb {
					if rgb[k] < 0 {
						rgb[k] = 0
					}
				}

				col.Add(rgb)

			}

			//fmt.Printf("col: %v\n", col)
		}

		sc.ReleaseRay(ray)
		sc.ReleaseShaderContext(chsc)

		col.Scale(1.0 / float32(totalSamples))

	} else {
		// Probabilistically decide whether to take BSDF or not.
		if cap(sc.Lsamples) < sc.NSamples {
			sc.Lsamples = make([]LightSample, 0, sc.NSamples)
		} else {
			sc.Lsamples = sc.Lsamples[:0]
		}

		sc.Lp.SampleArea(sc, sc.NSamples)

		for i := 0; i < sc.NSamples-len(bsdfSamples); i++ {

		}

		for _, ls := range sc.Lsamples {
			ray := sc.NewRay()
			chsc := sc.NewShaderContext()

			if m.Vec3Dot(ls.Ld, sc.Ng) < 0 {
				ray.Init(RayTypeShadow, sc.OffsetP(-1), m.Vec3Scale(ls.Ldist*(1.0-ShadowRayEpsilon), ls.Ld), 1.0, 0, sc)
			} else {
				ray.Init(RayTypeShadow, sc.OffsetP(1), m.Vec3Scale(ls.Ldist*(1.0-ShadowRayEpsilon), ls.Ld), 1.0, 0, sc)

			}

			if !TraceProbe(ray, chsc) {

				rho := bsdf.Eval(ls.Ld)

				//fmt.Printf("%v %v %v : \n", rho, sc.Liu, sc.Weight)

				rho.Mul(ls.Liu)
				rho.Scale(1.0 / ls.Weight)

				//fmt.Printf("%v\n\n", rho)
				rgb := rho.ToRGB()

				col.Add(rgb)

				sc.ReleaseRay(ray)
				sc.ReleaseShaderContext(chsc)
			}
		}

	}

	return col
}
