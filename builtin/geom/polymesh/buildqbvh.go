// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package polymesh

import (
	"errors"
	//"fmt"
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

			V0 := m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[i*3+0])+mesh.Verts.ElemsPerKey*key],
				mesh.Verts.Elems[int(mesh.idxp[i*3+0])+mesh.Verts.ElemsPerKey*key2], time)

			V1 := m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[i*3+1])+mesh.Verts.ElemsPerKey*key],
				mesh.Verts.Elems[int(mesh.idxp[i*3+1])+mesh.Verts.ElemsPerKey*key2], time)

			V2 := m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[i*3+2])+mesh.Verts.ElemsPerKey*key],
				mesh.Verts.Elems[int(mesh.idxp[i*3+2])+mesh.Verts.ElemsPerKey*key2], time)

			boxes[i].Reset()

			boxes[i].GrowVec3(V0)
			boxes[i].GrowVec3(V1)
			boxes[i].GrowVec3(V2)

			for k := 0; k < 3; k++ {
				centroids[i][0] = (V0[0] + V1[0] + V2[0]) / 3
			}

			idxs[i] = int32(i)
		}

		nodes, _ := qbvh.BuildAccelMotion(boxes, centroids, idxs, 16)
		mesh.accel.mqbvh.Nodes = nodes
		mesh.accel.idx = idxs

		return mesh.initMotionBoxes()

	}

	// Static mesh:

	for i := 0; i < mesh.facecount; i++ {

		V0 := mesh.Verts.Elems[int(mesh.idxp[i*3+0])]
		V1 := mesh.Verts.Elems[int(mesh.idxp[i*3+1])]
		V2 := mesh.Verts.Elems[int(mesh.idxp[i*3+2])]

		boxes[i].Reset()

		boxes[i].GrowVec3(V0)
		boxes[i].GrowVec3(V1)
		boxes[i].GrowVec3(V2)

		for k := 0; k < 3; k++ {
			centroids[i][0] = (V0[0] + V1[0] + V2[0]) / 3
		}

		if V0 == V1 && V0 == V2 {
			//log.Printf("nil triangle: %v %v %v %v %v\n", i, boxes[i], mesh.idxp[i*3+0], mesh.idxp[i*3+1], mesh.idxp[i*3+2])
		}
		idxs[i] = int32(i)
	}

	nodes, bounds := qbvh.BuildAccel(boxes, centroids, idxs, 16)
	mesh.accel.qbvh = nodes
	mesh.accel.idx = idxs
	mesh.bounds = bounds

	idxp := make([]uint32, len(mesh.idxp))

	for i := range idxs {
		idxp[i*3+0] = mesh.idxp[idxs[i]*3+0]
		idxp[i*3+1] = mesh.idxp[idxs[i]*3+1]
		idxp[i*3+2] = mesh.idxp[idxs[i]*3+2]
	}

	if mesh.uvtriidx != nil {

		uvidx := make([]uint32, len(mesh.uvtriidx))

		for i := range idxs {
			uvidx[i*3+0] = mesh.uvtriidx[idxs[i]*3+0]
			uvidx[i*3+1] = mesh.uvtriidx[idxs[i]*3+1]
			uvidx[i*3+2] = mesh.uvtriidx[idxs[i]*3+2]
		}
		mesh.uvtriidx = uvidx
	}

	if mesh.normalidx != nil {
		normalidx := make([]uint32, len(mesh.normalidx))

		for i := range idxs {
			normalidx[i*3+0] = mesh.normalidx[idxs[i]*3+0]
			normalidx[i*3+1] = mesh.normalidx[idxs[i]*3+1]
			normalidx[i*3+2] = mesh.normalidx[idxs[i]*3+2]
		}
		mesh.normalidx = normalidx
	}

	mesh.idxp = idxp

	/*
		fmt.Printf("Walk %v\n", mesh.Name())
		nodef := func(i int, b m.BoundingBox) {
			fmt.Printf(" nodes[%v] %v %v %v %v %v\n", i, b, nodes[i].Children[0], nodes[i].Children[1], nodes[i].Children[2], nodes[i].Children[3])
		}
		leaf := func(b m.BoundingBox, base, count int, empty bool) {
			fmt.Printf(" leaf: %v %v %v %v\n", b, base, count, empty)
		}

		qbvh.Walk(nodes, 0, nodef, leaf)
	*/
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

		mesh.motionBounds = append(mesh.motionBounds, box)
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
			leafBase := qbvh.LeafBase(mesh.accel.mqbvh.Nodes[node].Children[k])
			leafCount := qbvh.LeafCount(mesh.accel.mqbvh.Nodes[node].Children[k])

			for i := leafBase; i < leafBase+leafCount; i++ {
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
