// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package polymesh

import (
	"errors"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/qbvh"
)

func (mesh *PolyMesh) initAccel() error {
	boxes := make([]m.BoundingBox, mesh.facecount)
	centroids := make([]m.Vec3, mesh.facecount)
	idxs := make([]int32, mesh.facecount)

	if mesh.Verts.MotionKeys > 1 {
		k := 0.5 * float32(mesh.Verts.MotionKeys-1)

		time := k - m.Floor(k)

		key := int(m.Floor(k))
		key2 := int(m.Ceil(k))
		// if key == key2 then no interpolation needed but probably still do it to avoid branch

		for i := 0; i < mesh.facecount; i++ {
			var face Face

			face.V[0] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[i*3+0])+mesh.Verts.ElemsPerKey*key],
				mesh.Verts.Elems[int(mesh.idxp[i*3+0])+mesh.Verts.ElemsPerKey*key2], time)

			face.V[1] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[i*3+1])+mesh.Verts.ElemsPerKey*key],
				mesh.Verts.Elems[int(mesh.idxp[i*3+1])+mesh.Verts.ElemsPerKey*key2], time)

			face.V[2] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[i*3+2])+mesh.Verts.ElemsPerKey*key],
				mesh.Verts.Elems[int(mesh.idxp[i*3+2])+mesh.Verts.ElemsPerKey*key2], time)

			boxes[i] = face.Bounds()
			centroids[i] = face.Centroid()

			idxs[i] = int32(i)
		}

		nodes, _ := qbvh.BuildAccelMotion(boxes, centroids, idxs, 16)
		mesh.accel.mqbvh.Nodes = nodes
		mesh.accel.idx = idxs

		return mesh.initMotionBoxes()

	} else {
		for i := 0; i < mesh.facecount; i++ {
			var face Face

			face.V[0] = mesh.Verts.Elems[int(mesh.idxp[i*3+0])]
			face.V[1] = mesh.Verts.Elems[int(mesh.idxp[i*3+1])]
			face.V[2] = mesh.Verts.Elems[int(mesh.idxp[i*3+2])]

			boxes[i] = face.Bounds()
			centroids[i] = face.Centroid()

			idxs[i] = int32(i)
		}

		nodes, bounds := qbvh.BuildAccel(boxes, centroids, idxs, 16)
		mesh.accel.qbvh = nodes
		mesh.accel.idx = idxs
		mesh.bounds = bounds
		return nil

	}

	return nil
}

// assume the mesh.mqbvh nodes are already initialized with the tree topology
// For each motion key traverse the tree and initialize the leaf
func (mesh *PolyMesh) initMotionBoxes() error {

	if mesh.Verts.MotionKeys < 2 {
		return errors.New("initMotionBoxes: can't init with < 2 motion keys")
	}

	mesh.accel.mqbvh.Boxes = make([][]qbvh.MotionNodeBoxes, mesh.Verts.MotionKeys)

	var fullbox m.BoundingBox
	fullbox.Reset()

	for k := range mesh.accel.mqbvh.Boxes {
		mesh.accel.mqbvh.Boxes[k] = make([]qbvh.MotionNodeBoxes, len(mesh.accel.mqbvh.Nodes))
		box := mesh.initMotionBoxesRec(k, 0)
		fullbox.GrowBox(box)
	}

	mesh.bounds = fullbox
	return nil
}

func (mesh *PolyMesh) initMotionBoxesRec(key int, node int32) (nodebox m.BoundingBox) {

	nodebox.Reset()

	for k := range mesh.accel.mqbvh.Nodes[node].Children {
		if mesh.accel.mqbvh.Nodes[node].Children[k] < 0 {

			if mesh.accel.mqbvh.Nodes[node].Children[k] == -1 {
				//Empty leaf
				continue
			}
			var box m.BoundingBox

			box.Reset()

			// calc box
			leaf_base := qbvh.LEAF_BASE(mesh.accel.mqbvh.Nodes[node].Children[k])
			leaf_count := qbvh.LEAF_COUNT(mesh.accel.mqbvh.Nodes[node].Children[k])

			for i := leaf_base; i < leaf_base+leaf_count; i++ {
				faceidx := int(mesh.accel.idx[i])

				box.GrowVec3(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+0])+(mesh.Verts.ElemsPerKey*key)])
				box.GrowVec3(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+1])+(mesh.Verts.ElemsPerKey*key)])
				box.GrowVec3(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+2])+(mesh.Verts.ElemsPerKey*key)])
			}

			//mesh.mqbvh.Nodes[node].SetBounds(k, box)
			mesh.accel.mqbvh.Boxes[key][node].SetBounds(k, box)

			nodebox.GrowBox(box)

		} else if mesh.accel.mqbvh.Nodes[node].Children[k] >= 0 {
			box := mesh.initMotionBoxesRec(key, mesh.accel.mqbvh.Nodes[node].Children[k])
			mesh.accel.mqbvh.Boxes[key][node].SetBounds(k, box)
			nodebox.GrowBox(box)
		}

	}

	return
}
