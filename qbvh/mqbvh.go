// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qbvh

import (
	m "github.com/jamiec7919/vermeer/math"
)

// 96 byte struct (24*4, 1 1/12 cache lines)
// Each motion key will have one set of boxes
type MotionNodeBoxes [4 * 3 * 2]float32

// 32 bytes
type MotionNode struct {
	Axis0, Axis1, Axis2 int32
	Children            [4]int32
	Parent              int32
	pad                 [2]uint32
}

// There will be one set of boxes per pair of motion knots.  Topology of
// tree is same for all trees. Boxes are interpolated before intersection.
type MotionQBVH struct {
	Boxes [][]MotionNodeBoxes
	Nodes []MotionNode
}

func (n *MotionNodeBoxes) SetBounds(idx int, bounds m.BoundingBox) {
	//log.Printf("SetBounds %v %v", idx, bounds)
	for i := 0; i < 2; i++ {
		for k := 0; k < 3; k++ {
			n[idx+(i*12)+(k*4)] = bounds.Bounds[i][k]
		}
	}
	//log.Printf("SetBounds %v %v", idx, n.boxes)
}

func (n *MotionNodeBoxes) Bounds(idx int) (bounds m.BoundingBox) {
	for i := 0; i < 2; i++ {
		for k := 0; k < 3; k++ {
			bounds.Bounds[i][k] = n[idx+(i*12)+(k*4)]
		}
	}

	return
}

func (n *MotionNode) SetEmptyLeaf(idx int) {
	n.Children[idx] = -1
}

// count <= 16
func (n *MotionNode) SetLeaf(idx int, first, count uint32) {
	if count == 0 {
		n.SetEmptyLeaf(idx)
		return
	}

	v := (1 << 31) | ((first << 4) & 0xfffffff0) | ((count - 1) & 0xf)
	n.Children[idx] = int32(v)
	//log.Printf("%v %v", n.children[idx], v)
	//log.Printf("%v %v", (n.children[idx]&(0x7fffffff))>>4, (v&0x7fffffff)>>4)
	//log.Printf("%v %v", n.children[idx]&0xf, v&0xf)
}

func (n *MotionNode) SetChild(idx int, ch int32) {
	n.Children[idx] = ch
}

func (n *MotionNode) SetAxis0(axis int32) {
	n.Axis0 = axis
}

func (n *MotionNode) SetAxis1(axis int32) {
	n.Axis1 = axis
}

func (n *MotionNode) SetAxis2(axis int32) {
	n.Axis2 = axis
}
