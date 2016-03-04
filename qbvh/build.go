// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qbvh

/* Will try Early Split Clipping (2007 Ernst/Greiner).

In this approach each input triangle is tested for overall surface area, if too large it is split and
boxes recomputed recursively.  Geometry is NOT CHANGED, BVH must store multiple references to the
individual triangle.  This does mean that we'll either need to duplicate triangles or reinstitute
triangle refs for meshes.  For grids this probably wont be necessary as we compute a fine tesselation anyway
but may still be beneficial for some.

*/

import (
	"fmt"
	m "github.com/jamiec7919/vermeer/math"
)

const nbins = 8

//TODO: is this necessary?
//var testLeafs []int32

// Returns the (axis,pivot) and the indxs array sorted
func calcMinCost(bounds m.BoundingBox, centroids []m.Vec3, boxes []m.BoundingBox, indxs []int32) (int, int) {
	var binBounds [nbins]m.BoundingBox
	var binN [nbins]int32

	axis := bounds.MaxDim()

	// A flat bounding box, this should only ever occur with 1 triangle.??
	if bounds.Bounds[1][axis] == bounds.Bounds[0][axis] {

		if len(indxs) > 1 {
			panic(fmt.Sprintf("calcMinCost: null axis %v %v %v: len(indxs)>1 (%v) %v %v", axis, bounds.Bounds[1][axis]-bounds.Bounds[0][axis], bounds, len(indxs), indxs, boxes))
		}
		return axis, len(indxs)/2 + 1
	}

	k1 := float32(nbins) * (1.0 - 0.00006) / (bounds.Bounds[1][axis] - bounds.Bounds[0][axis])
	k0 := bounds.Bounds[0][axis]

	for i := range binBounds {
		binBounds[i].Reset()
	}

	for i := range indxs {
		//c := boxes[indxs[i]].AxisCentroid(axis)
		c := centroids[i][axis]
		bin := int(k1 * (c - k0))

		if bin < 0 {
			panic(fmt.Sprintf("calcMinCost: bin < 0 (%v) %v %v %v %v", len(indxs), bin, c, k0, k1))
		}
		if bin > nbins-1 {
			panic(fmt.Sprintf("calcMinCost: bin > Nbins-1 (%v) %v %v %v %v", len(indxs), bin, c, k0, k1))
		}

		binN[bin]++
		binBounds[bin].GrowBox(boxes[i])
	}

	// Evaluate costs from the left then the right
	var lbox [nbins]m.BoundingBox
	var rbox [nbins]m.BoundingBox
	var lN [nbins]int32
	var rN [nbins]int32

	var box m.BoundingBox
	box.Reset()

	n := int32(0)

	for i := 0; i < nbins; i++ {
		box.GrowBox(binBounds[i])
		n += binN[i]
		lbox[i] = box
		lN[i] = n
	}

	box.Reset()
	n = 0

	for i := 0; i < nbins; i++ {
		box.GrowBox(binBounds[nbins-1-i])
		n += binN[nbins-1-i]
		rbox[nbins-1-i] = box
		rN[nbins-1-i] = n
	}

	// Can now consider l[i] & r[i] as accumulations
	binMinCost := -1
	minCost := m.Inf(1)
	for i := 1; i < nbins; i++ {
		cost := lbox[i-1].SurfaceArea()*float32(lN[i-1]) + rbox[i].SurfaceArea()*float32(rN[i])
		//fmt.Printf("cost(%v) %v %v %v+%v %v->%v\n", len(indxs), cost, i, lN[i-1], rN[i], lbox[i-1].SurfaceArea(), rbox[i].SurfaceArea())
		if cost < minCost {
			//		log.Printf("cost< %v %v", cost, i)
			binMinCost = i
			minCost = cost
		}
	}

	//	log.Printf("%v bin %v,%v", binMinCost, lN[binMinCost-1], rN[binMinCost])

	left := 0
	right := len(indxs) - 1
	//log.Printf("a %v", indxs)
	for left <= right {
		//		c := boxes[indxs[left]].AxisCentroid(axis)
		c := centroids[left][axis]

		bin := int(k1 * (c - k0))
		//log.Printf("%v %v %v", indxs[left], bin, bin < binMinCost)
		if bin < binMinCost {
			left++
		} else {
			indxs[left], indxs[right] = indxs[right], indxs[left]
			centroids[left], centroids[right] = centroids[right], centroids[left]
			boxes[left], boxes[right] = boxes[right], boxes[left]
			right--
		}
		//log.Printf("%v %v", left, indxs)

	}

	//log.Printf("b(%v) %v", left, indxs)
	//	log.Printf("%v %v", left, right)

	return axis, left
}

func calcBox(boxes []m.BoundingBox, indxs []int32) (box m.BoundingBox) {
	box.Reset()
	//if len(indxs) == 0 {
	//	box.Grow(0, 0, 0)
	//	return
	//	}
	for i := range indxs {
		box.GrowBox(boxes[i])
	}

	return
}

func binarySplit(boxes []m.BoundingBox, centroids []m.Vec3, indxs []int32) (axis, pivot int, bounds m.BoundingBox) {

	bounds.Reset()

	for i := range indxs {
		/*		P0 := ms.Vp[ms.F[indxs[i]*3+0]]
				P1 := ms.Vp[ms.F[indxs[i]*3+1]]
				P2 := ms.Vp[ms.F[indxs[i]*3+2]]
				ci := m.Vec3{
					(1.0 / 3.0) * (P0[0] + P1[0] + P2[0]),
					(1.0 / 3.0) * (P0[1] + P1[1] + P2[1]),
					(1.0 / 3.0) * (P0[2] + P1[2] + P2[2]),
				}
		*/
		bounds.GrowVec3(centroids[i])
		//bounds.GrowVec3(ci)
		//		bounds.GrowVec3(boxes[indxs[i]].Centroid())
	}

	if len(indxs) <= 4 {
		axis = 0
		pivot = len(indxs)
		return
	}
	axis, pivot = calcMinCost(bounds, centroids, boxes, indxs)

	return
}

func buildAccelRec(nodes *[]Node, boxes []m.BoundingBox, centroids []m.Vec3, indxs []int32, leafMax, baseidx int) (int32, m.BoundingBox) {
	//log.Printf("A %v", len(indxs))
	axis0, pivot0, _ := binarySplit(boxes, centroids, indxs)

	if len(indxs[:pivot0]) < 17 {
		// create leaf
	}

	//log.Printf("B %v %v", pivot0, len(indxs[:pivot0]))
	axis1, pivot1, _ := binarySplit(boxes[:pivot0], centroids[:pivot0], indxs[:pivot0])
	//log.Printf("C")
	axis2, pivot2, _ := binarySplit(boxes[pivot0:], centroids[pivot0:], indxs[pivot0:])

	//log.Printf("Node %v %v %v %v %v %v %v %v %v", axis0, pivot0, bounds0, axis1, pivot1, bounds1, axis2, pivot2, bounds2)
	nodei := int32(len(*nodes))
	*nodes = append(*nodes, Node{})

	(*nodes)[nodei].SetAxis0(int32(axis0))
	(*nodes)[nodei].SetAxis1(int32(axis1))
	(*nodes)[nodei].SetAxis2(int32(axis2))

	if len(indxs[:pivot1]) <= leafMax {
		//	log.Printf("Leaf0 %v", indxs[:pivot1
		(*nodes)[nodei].SetBounds(0, calcBox(boxes[:pivot1], indxs[:pivot1]))
		(*nodes)[nodei].SetLeaf(0, uint32(baseidx), uint32(len(indxs[:pivot1])))
		/*
			if uint32(len(indxs[:pivot1])) < 1 {
				//log.Printf("Empty leaf")
			}

			for k := range indxs[:pivot1] {
				testLeafs[indxs[:pivot1][k]] = int32(baseidx)
			}
		*/
	} else {
		child, box := buildAccelRec(nodes, boxes[:pivot1], centroids[:pivot1], indxs[:pivot1], leafMax, baseidx)
		(*nodes)[nodei].SetBounds(0, box)
		(*nodes)[nodei].SetChild(0, child)

	}

	if len(indxs[pivot1:pivot0]) <= leafMax {
		//	log.Printf("Leaf1 %v", indxs[pivot1:pivot0])
		(*nodes)[nodei].SetBounds(1, calcBox(boxes[pivot1:pivot0], indxs[pivot1:pivot0]))
		(*nodes)[nodei].SetLeaf(1, uint32(baseidx+pivot1), uint32(len(indxs[pivot1:pivot0])))
		/*
			if uint32(len(indxs[pivot1:pivot0])) < 1 {
				//log.Printf("Empty leaf")
			}
			for k := range indxs[pivot1:pivot0] {
				testLeafs[indxs[pivot1:pivot0][k]] = int32(baseidx + pivot1)
			}
		*/
	} else {
		child, box := buildAccelRec(nodes, boxes[pivot1:pivot0], centroids[pivot1:pivot0], indxs[pivot1:pivot0], leafMax, baseidx+pivot1)
		(*nodes)[nodei].SetBounds(1, box)
		(*nodes)[nodei].SetChild(1, child)
	}

	if len(indxs[pivot0:pivot0+pivot2]) <= leafMax {
		//	log.Printf("Leaf2 %v", indxs[pivot0:pivot0+pivot2])
		(*nodes)[nodei].SetBounds(2, calcBox(boxes[pivot0:pivot0+pivot2], indxs[pivot0:pivot0+pivot2]))
		(*nodes)[nodei].SetLeaf(2, uint32(baseidx+pivot0), uint32(len(indxs[pivot0:pivot0+pivot2])))
		/*
			if uint32(len(indxs[pivot0:pivot0+pivot2])) < 1 {
				//log.Printf("Empty leaf")
			}
			for k := range indxs[pivot0 : pivot0+pivot2] {
				testLeafs[indxs[pivot0 : pivot0+pivot2][k]] = int32(baseidx + pivot0)
			}
		*/
	} else {
		child, box := buildAccelRec(nodes, boxes[pivot0:pivot0+pivot2], centroids[pivot0:pivot0+pivot2], indxs[pivot0:pivot0+pivot2], leafMax, baseidx+pivot0)
		(*nodes)[nodei].SetBounds(2, box)
		(*nodes)[nodei].SetChild(2, child)
	}

	if len(indxs[pivot0+pivot2:]) <= leafMax {
		//	log.Printf("Leaf3 %v", indxs[pivot0+pivot2:])
		(*nodes)[nodei].SetBounds(3, calcBox(boxes[pivot0+pivot2:], indxs[pivot0+pivot2:]))
		(*nodes)[nodei].SetLeaf(3, uint32(baseidx+pivot0+pivot2), uint32(len(indxs[pivot0+pivot2:])))
		/*
			if uint32(len(indxs[pivot0+pivot2:])) < 1 {
				//log.Printf("Empty leaf")
			}
				for k := range indxs[pivot0+pivot2:] {
					testLeafs[indxs[pivot0+pivot2:][k]] = int32(baseidx + pivot0 + pivot2)
				}
		*/
	} else {
		child, box := buildAccelRec(nodes, boxes[pivot0+pivot2:], centroids[pivot0+pivot2:], indxs[pivot0+pivot2:], leafMax, baseidx+pivot0+pivot2)
		(*nodes)[nodei].SetBounds(3, box)
		(*nodes)[nodei].SetChild(3, child)
	}

	//TODO: sum of children!
	nodebox := m.BoundingBox{}
	nodebox.Reset()
	for i := 0; i < 4; i++ {
		//nodebox.GrowBox(boxes[indxs[i]])
		nodebox.GrowBox((*nodes)[nodei].Bounds(i))
	}

	return nodei, nodebox
}

// This returns a slice of allocated nodes (node 0 is root) and ALSO sorts the indxs slice to match the leaf structure
// of the node tree. leafMax must be <= 16
func BuildAccel(boxes []m.BoundingBox, centroids []m.Vec3, indxs []int32, leafMax int) (nodes []Node, bounds m.BoundingBox) {
	if leafMax > 16 {
		leafMax = 16
	}

	if leafMax < 1 {
		leafMax = 1
	}

	//testLeafs = make([]int32, len(indxs))

	_, bounds = buildAccelRec(&nodes, boxes, centroids, indxs, leafMax, 0)

	return
}
