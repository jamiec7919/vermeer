// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/jamiec7919/vermeer/colour"
	m "github.com/jamiec7919/vermeer/math"
)

// TraceSample is returned by Trace.
type TraceSample struct {
	Colour  colour.RGB
	Opacity colour.RGB
	Alpha   float32
	Point   m.Vec3
	Z       float64
	ElemID  uint32
	Geom    Geom
}

// TraceProbe intersects ray with the scene and sets up the globals sg with the first intersection.
// rays of type RayTypeShadow will early-out and not necessarily return the first intersection.
// Returns true if any intersection or false for none.
func TraceProbe(ray *Ray, sg *ShaderContext) bool {

	stats.incRayCount()

	if ray.Type&RayTypeShadow != 0 {
		stats.incShadowRayCount()
	}

	return scene.Trace(ray, sg)
}

// Trace intersects ray with the scene and evaluates the shader at the first intersection. The
// result is returned in the samp struct.
// Returns true if any intersection or false for none.
func Trace(ray *Ray, samp *TraceSample) bool {

	// This is the only time that ShaderContext should be created manually, note we set task here.
	sg := &ShaderContext{
		Ro:           ray.P,
		Rd:           ray.D,
		X:            ray.X,
		Y:            ray.Y,
		Sx:           ray.Sx,
		Sy:           ray.Sy,
		Level:        ray.Level,
		Lambda:       ray.Lambda,
		I:            ray.I,
		Time:         ray.Time,
		task:         ray.Task,
		Image:        image,
		Scramble:     ray.Scramble,
		Transform:    m.Matrix4Identity(),
		InvTransform: m.Matrix4Identity(),
	}

	Tclosest := ray.Tclosest // record the original value for this as it gets overwritten

retry:

	if TraceProbe(ray, sg) {

		if sg.Shader == nil { // can't do much with no material
			return false
		}

		// set interior list for shader
		sg.InteriorList = ray.InteriorList

		// Set ray length
		sg.Rl = float64(ray.Tclosest)

		ray.DifferentialTransfer(sg)

		sg.ApplyTransform()

		// If this shader/point won't evaluate this intersection then try to find the
		// next one.
		if !sg.Shader.Eval(sg) {
			ray.Tmin = ray.Tclosest + 0.0001
			ray.Tclosest = Tclosest // Reset the ray
			// Transfer current interior list to ray (sg.Eval will probably have updated it).
			ray.InteriorList = sg.InteriorList
			goto retry
		}

		if samp != nil {
			samp.Colour = sg.OutRGB
			samp.Point = sg.P
			samp.ElemID = sg.ElemID
			samp.Geom = sg.Geom
		}

		return true
	}
	return false
}
