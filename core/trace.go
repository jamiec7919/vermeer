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
		Ro:     ray.P,
		Rd:     ray.D,
		X:      ray.X,
		Y:      ray.Y,
		Sx:     ray.Sx,
		Sy:     ray.Sy,
		Level:  ray.Level,
		Lambda: ray.Lambda,
		I:      ray.I,
		Time:   ray.Time,
		task:   ray.Task,
	}

	if TraceProbe(ray, sg) {

		if sg.Shader == nil { // can't do much with no material
			return false
		}

		sg.Shader.Eval(sg)

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
