// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qbvh

import (
	//	"fmt"
	m "github.com/jamiec7919/vermeer/math"
)

func buildAccelMotionRec(nodes *[]MotionNode, boxes []m.BoundingBox, centroids []m.Vec3, indxs []int32, leafMax, baseidx int) int32 {
	//log.Printf("A %v", len(indxs))
	axis0, pivot0, _ := binarySplit(boxes, centroids, leafMax, indxs)

	if len(indxs[:pivot0]) < 17 {
		// create leaf
	}

	//log.Printf("B %v %v", pivot0, len(indxs[:pivot0]))
	axis1, pivot1, _ := binarySplit(boxes[:pivot0], centroids[:pivot0], leafMax, indxs[:pivot0])
	//log.Printf("C")
	axis2, pivot2, _ := binarySplit(boxes[pivot0:], centroids[pivot0:], leafMax, indxs[pivot0:])

	//log.Printf("Node %v %v %v %v %v %v %v %v %v", axis0, pivot0, bounds0, axis1, pivot1, bounds1, axis2, pivot2, bounds2)
	nodei := int32(len(*nodes))
	*nodes = append(*nodes, MotionNode{})

	(*nodes)[nodei].SetAxis0(int32(axis0))
	(*nodes)[nodei].SetAxis1(int32(axis1))
	(*nodes)[nodei].SetAxis2(int32(axis2))

	if len(indxs[:pivot1]) <= leafMax {
		//	log.Printf("Leaf0 %v", indxs[:pivot1
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
		child := buildAccelMotionRec(nodes, boxes[:pivot1], centroids[:pivot1], indxs[:pivot1], leafMax, baseidx)
		(*nodes)[nodei].SetChild(0, child)

	}

	if len(indxs[pivot1:pivot0]) <= leafMax {
		//	log.Printf("Leaf1 %v", indxs[pivot1:pivot0])
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
		child := buildAccelMotionRec(nodes, boxes[pivot1:pivot0], centroids[pivot1:pivot0], indxs[pivot1:pivot0], leafMax, baseidx+pivot1)
		(*nodes)[nodei].SetChild(1, child)
	}

	if len(indxs[pivot0:pivot0+pivot2]) <= leafMax {
		//	log.Printf("Leaf2 %v", indxs[pivot0:pivot0+pivot2])
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
		child := buildAccelMotionRec(nodes, boxes[pivot0:pivot0+pivot2], centroids[pivot0:pivot0+pivot2], indxs[pivot0:pivot0+pivot2], leafMax, baseidx+pivot0)
		(*nodes)[nodei].SetChild(2, child)
	}

	if len(indxs[pivot0+pivot2:]) <= leafMax {
		//	log.Printf("Leaf3 %v", indxs[pivot0+pivot2:])
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
		child := buildAccelMotionRec(nodes, boxes[pivot0+pivot2:], centroids[pivot0+pivot2:], indxs[pivot0+pivot2:], leafMax, baseidx+pivot0+pivot2)
		(*nodes)[nodei].SetChild(3, child)
	}

	return nodei
}

// BuildAccelMotion constructs a MQBVH tree for the given elements.  Elements are described by
// bounding boxes, centroids (which may be different from the box centroid) and an index array.
// This returns a slice of allocated nodes (node 0 is root) and ALSO sorts the indxs slice to match the leaf structure
// of the node tree. leafMax must be <= 16.
// Note that the bounding boxes for internal nodes must be recalculated for each key as only the
// tree node topology is constructed by BuildAccelMotion.
func BuildAccelMotion(boxes []m.BoundingBox, centroids []m.Vec3, indxs []int32, leafMax int) (nodes []MotionNode, bounds m.BoundingBox) {
	if leafMax > 16 {
		leafMax = 16
	}

	if leafMax < 1 {
		leafMax = 1
	}

	//testLeafs = make([]int32, len(indxs))

	buildAccelMotionRec(&nodes, boxes, centroids, indxs, leafMax, 0)

	return
}
