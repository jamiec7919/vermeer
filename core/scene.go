// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/qbvh"
)

type Scene struct {
	prims  []Primitive
	nodes  []qbvh.Node
	bounds m.BoundingBox

	lights []Light
}

func (s *Scene) VisRay(ray *RayData) {
	s.visRayAccel(ray)

}

func (s *Scene) TraceRay(ray *RayData) {
	s.traceRayAccel(ray)

}

func (scene *Scene) initAccel() error {
	boxes := make([]m.BoundingBox,0, len(scene.prims))
	indices := make([]int32, 0,len(scene.prims))
	centroids := make([]m.Vec3,0, len(scene.prims))

	for i := range scene.prims {
		if !scene.prims[i].Visible() {continue}
		box :=scene.prims[i].WorldBounds()
		boxes = append(boxes,box)
		indices = append(indices,int32(i))
		centroids = append(centroids,box.Centroid())
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
