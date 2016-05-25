// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/jamiec7919/vermeer/colour"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/qbvh"
)

type Scene struct {
	prims  []Primitive
	nodes  []qbvh.Node
	bounds m.BoundingBox

	lights []Light
}

var grc *RenderContext

type ScreenSample struct {
	Colour  colour.RGB
	Opacity colour.RGB
	Alpha   float32
	Point   m.Vec3
	Z       float64
	ElemId  uint32
	Prim    Primitive
}

func TraceProbe(ray *RayData, sg *ShaderGlobals) bool {

	if ray.Type&RAY_SHADOW != 0 {
		grc.scene.visRayAccel(ray)
		shadowRays++
		return !ray.IsVis()
	}

	mtlid := grc.scene.traceRayAccel(ray, sg)

	if mtlid != -1 {

		mtl := GetMaterial(mtlid)

		sg.Shader = mtl
		sg.N = m.Vec3Normalize(sg.N)
		sg.Ns = m.Vec3Normalize(sg.Ns)
		return true
	}

	return false
}

func Trace(ray *RayData, samp *ScreenSample) bool {
	rayCount++
	sg := &ShaderGlobals{
		Ro:     ray.Ray.P,
		Rd:     ray.Ray.D,
		Prim:   ray.Result.Prim,
		ElemId: ray.Result.ElemId,
		Depth:  ray.Level,
		rnd:    ray.rnd,
		Lambda: ray.Lambda,
		Time:   ray.Time,
	}

	if TraceProbe(ray, sg) {
		if sg.Shader == nil { // can't do much with no material
			return false
		}

		sg.Shader.Eval(sg)

		if samp != nil {
			samp.Colour = sg.OutRGB
			samp.Point = sg.Ro
			samp.ElemId = sg.ElemId
			samp.Prim = sg.Prim
		}

		return true
	}
	return false
}

func (scene *Scene) initAccel() error {
	boxes := make([]m.BoundingBox, 0, len(scene.prims))
	indices := make([]int32, 0, len(scene.prims))
	centroids := make([]m.Vec3, 0, len(scene.prims))

	for i := range scene.prims {
		if !scene.prims[i].Visible() {
			continue
		}
		box := scene.prims[i].WorldBounds()
		boxes = append(boxes, box)
		indices = append(indices, int32(i))
		centroids = append(centroids, box.Centroid())
	}

	nodes, bounds := qbvh.BuildAccel(boxes, centroids, indices, 1)

	scene.nodes = nodes

	// Rearrange (visible) primitive array to match leaf structure
	nprims := make([]Primitive, len(indices))

	for i := range indices {
		nprims[i] = scene.prims[indices[i]]
	}

	scene.prims = nprims
	scene.bounds = bounds

	return nil
}
