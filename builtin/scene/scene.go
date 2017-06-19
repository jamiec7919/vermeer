// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scene

import (
	"errors"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/qbvh"
)

// Scene is the default scene type.
type Scene struct {
	qbvh   []qbvh.Node
	mqbvh  qbvh.MotionQBVH
	geoms  []core.Geom
	lights []core.Light
	bounds m.BoundingBox
}

// New returns a new empty scene.
func New() *Scene {
	return &Scene{}
}

// Trace returns true if ray hits something. Usually this is the first hit along the ray
// unless ray.Type &= RayTypeShadow.
func (s *Scene) Trace(ray *core.Ray, sg *core.ShaderContext) bool {
	/*	hit := false

		for i := range s.geoms {
			if s.geoms[i].Trace(ray, sg) {
				hit = true

				if ray.Type&core.RayTypeShadow != 0 {
					return true
				}
			}

		}
		return hit*/

	if s.qbvh != nil {
		return qbvh.Trace(s.qbvh, s, ray, sg)
	} else {
		// Motion
		k := ray.Time * float32(len(s.mqbvh.Boxes)-1)

		time := k - m.Floor(k)

		key := int(m.Floor(k))
		key2 := int(m.Ceil(k))

		return qbvh.TraceMotion(s.mqbvh, time, key, key2, s, ray, sg)
	}
}

// TraceElems implements qbvh.Geom.
func (s *Scene) TraceElems(ray *core.Ray, sc *core.ShaderContext, base, count int) bool {

	hit := false

	for i := base; i < base+count; i++ {
		if ray.Type&core.RayTypeShadow != 0 {
			// If this is a shadow ray but the geom is marked as non-opaque then can't
			// affect shadow vis.
			if !s.geoms[i].HasFlags(core.GeomFlagOpaque) {
				continue
			}
		}

		if s.geoms[i].Trace(ray, sc) {
			sc.Geom = s.geoms[i]

			hit = true

			if ray.Type&core.RayTypeShadow != 0 {
				return true
			}
		}
	}

	return hit
}

// TraceElemsMotion implements qbvh.MotionGeom.
func (s *Scene) TraceMotionElems(time float32, key, key2 int, ray *core.Ray, sc *core.ShaderContext, base, count int) bool {
	hit := false

	for i := base; i < base+count; i++ {
		if s.geoms[i].Trace(ray, sc) {
			sc.Geom = s.geoms[i]

			hit = true

			if ray.Type&core.RayTypeShadow != 0 {
				return true
			}
		}
	}

	return hit
}

// LightsPrepare returns the list of lights.
func (s *Scene) LightsPrepare(sg *core.ShaderContext) {

	if sg.Lights != nil {
		sg.Lights = sg.Lights[:0]
	}

	for _, l := range s.lights {

		geom := l.Geom()

		if geom != sg.Geom {
			sg.Lights = append(sg.Lights, l)

		}
	}

}

// AddGeom adds the Geom to the scene
func (s *Scene) AddGeom(geom core.Geom) error {
	s.geoms = append(s.geoms, geom)
	return nil
}

// AddLight adds the light to the scene
func (s *Scene) AddLight(light core.Light) error {
	s.lights = append(s.lights, light)
	return nil
}

// PreRender is called after all other nodes PreRender.
func (s *Scene) PreRender() error {
	return s.initAccel()
}

func (s *Scene) initAccel() error {

	boxes := make([]m.BoundingBox, 0, len(s.geoms))
	indices := make([]int32, 0, len(s.geoms))
	centroids := make([]m.Vec3, 0, len(s.geoms))

	maxKeys := 0

	for i := range s.geoms {
		if s.geoms[i].MotionKeys() > maxKeys {
			maxKeys = s.geoms[i].MotionKeys()
		}
	}

	if maxKeys == 1 {
		for i := range s.geoms {
			//		if !s.geoms[i].Visible() {
			//			continue
			//		}
			box := s.geoms[i].Bounds(0)
			boxes = append(boxes, box)
			indices = append(indices, int32(i))
			centroids = append(centroids, box.Centroid())
		}

		nodes, bounds := qbvh.BuildAccel(boxes, centroids, indices, 1)

		s.qbvh = nodes

		// Rearrange (visible) primitive array to match leaf structure
		ngeoms := make([]core.Geom, len(indices))

		for i := range indices {
			ngeoms[i] = s.geoms[indices[i]]
		}

		s.geoms = ngeoms
		s.bounds = bounds

	} else {

		for i := range s.geoms {
			//		if !s.geoms[i].Visible() {
			//			continue
			//		}
			box := s.geoms[i].Bounds(0.5)
			boxes = append(boxes, box)
			indices = append(indices, int32(i))
			centroids = append(centroids, box.Centroid())
		}

		nodes, _ := qbvh.BuildAccelMotion(boxes, centroids, indices, 1)

		s.mqbvh.Nodes = nodes

		// Rearrange (visible) primitive array to match leaf structure
		ngeoms := make([]core.Geom, len(indices))

		for i := range indices {
			ngeoms[i] = s.geoms[indices[i]]
		}

		s.geoms = ngeoms
		s.bounds.Reset()

		return s.initMotionBoxes(maxKeys)
	}
	return nil
}

// assume the s.mqbvh nodes are already initialized with the tree topology
// For each motion key traverse the tree and initialize the leaf
func (s *Scene) initMotionBoxes(keys int) error {

	if keys < 2 {
		return errors.New("s.initMotionBoxes: can't init with < 2 motion keys")
	}

	s.mqbvh.Boxes = make([][]qbvh.MotionNodeBoxes, keys)

	var fullbox m.BoundingBox
	fullbox.Reset()

	for k := range s.mqbvh.Boxes {
		s.mqbvh.Boxes[k] = make([]qbvh.MotionNodeBoxes, len(s.mqbvh.Nodes))
		box := s.initMotionBoxesRec(k, 0)
		fullbox.GrowBox(box)
	}

	s.bounds = fullbox
	return nil
}

func (s *Scene) initMotionBoxesRec(key int, node int32) (nodebox m.BoundingBox) {

	nodebox.Reset()

	for k := range s.mqbvh.Nodes[node].Children {
		if s.mqbvh.Nodes[node].Children[k] < 0 {

			if s.mqbvh.Nodes[node].Children[k] == -1 {
				//Empty leaf
				continue
			}
			var box m.BoundingBox

			box.Reset()

			// calc box
			leafBase := qbvh.LeafBase(s.mqbvh.Nodes[node].Children[k])
			leafCount := qbvh.LeafCount(s.mqbvh.Nodes[node].Children[k])

			for i := leafBase; i < leafBase+leafCount; i++ {

				time := float32(key) / float32(len(s.mqbvh.Boxes)-1)

				box.GrowBox(s.geoms[i].Bounds(time))
			}

			//mesh.mqbvh.Nodes[node].SetBounds(k, box)
			s.mqbvh.Boxes[key][node].SetBounds(k, box)

			nodebox.GrowBox(box)

		} else if s.mqbvh.Nodes[node].Children[k] >= 0 {
			box := s.initMotionBoxesRec(key, s.mqbvh.Nodes[node].Children[k])
			s.mqbvh.Boxes[key][node].SetBounds(k, box)
			nodebox.GrowBox(box)
		}

	}

	return
}
