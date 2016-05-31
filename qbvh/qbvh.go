// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package qbvh implements the Quad-BVH data structure.
*/
package qbvh

import (
	m "github.com/jamiec7919/vermeer/math"
)

// MaxLeafCount is the maximum leaf size supported by data structure.
// Users of QBVH may request smaller leafs.
const MaxLeafCount = 16

// Index is the type of a triangle/prim index.
type Index int32

// Node represenets a QBVH node.
// 128 byte struct
type Node struct {
	Boxes               [4 * 3 * 2]float32
	Axis0, Axis1, Axis2 int32
	Children            [4]int32
	Parent              int32
}

// SetBounds sets the node bounding box for the given child index.
func (n *Node) SetBounds(idx int, bounds m.BoundingBox) {
	//log.Printf("SetBounds %v %v", idx, bounds)
	for i := 0; i < 2; i++ {
		for k := 0; k < 3; k++ {
			n.Boxes[idx+(i*12)+(k*4)] = bounds.Bounds[i][k]
		}
	}
	//log.Printf("SetBounds %v %v", idx, n.boxes)
}

// Bounds returns the bounds for the given child.
func (n *Node) Bounds(idx int) (bounds m.BoundingBox) {
	for i := 0; i < 2; i++ {
		for k := 0; k < 3; k++ {
			bounds.Bounds[i][k] = n.Boxes[idx+(i*12)+(k*4)]
		}
	}

	return
}

// LEAF_COUNT returns the number of elements stored in leaf node l.
func LEAF_COUNT(l int32) int { return int((((l) & 0xf) + 1)) }

// LEAF_BASE returns the first index of thw elements stored in leaf node l.
func LEAF_BASE(l int32) int { return int(((l) & 0x7ffffff) >> 4) }

// SetEmptyLeaf sets the given child index to an empty leaf.
func (n *Node) SetEmptyLeaf(idx int) {
	n.Children[idx] = -1
	b := m.InfBox
	//b := m.BoundingBox{}
	b.Reset() // setting it like this will not stop rays hitting!
	/*	b.Bounds[0][0] = m.Inf(1)
		b.Bounds[0][1] = 50000
		b.Bounds[0][1] = 50000
		b.Bounds[1][0] = 50000
		b.Bounds[1][1] = 50000
		b.Bounds[1][1] = 50000*/
	n.SetBounds(idx, b)
}

// SetLeaf sets the given child index.  count must be <= MaxLeafCount (16).
func (n *Node) SetLeaf(idx int, first, count uint32) {
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

// SetChild sets the child node.
func (n *Node) SetChild(idx int, ch int32) {
	n.Children[idx] = ch
}

// SetAxis0 sets the split axis for first split.
func (n *Node) SetAxis0(axis int32) {
	n.Axis0 = axis
}

// SetAxis1 sets the split axis for left most split.
func (n *Node) SetAxis1(axis int32) {
	n.Axis1 = axis
}

// SetAxis2 sets the split axis for right most split.
func (n *Node) SetAxis2(axis int32) {
	n.Axis2 = axis
}

// Walk will recursively walk the tree and call the functions nodef and leaf for each node/leaf.
func Walk(nodes []Node, node int, nodef func(i int, bounds m.BoundingBox), leaf func(bounds m.BoundingBox, base, count int, empty bool)) {
	for i := range nodes[node].Children {
		if nodes[node].Children[i] < 0 {
			leaf(nodes[node].Bounds(i), LEAF_BASE(nodes[node].Children[i]), LEAF_COUNT(nodes[node].Children[i]), nodes[node].Children[i] == -1)
		} else {
			nodef(int(nodes[node].Children[i]), nodes[node].Bounds(i))
			Walk(nodes, int(nodes[node].Children[i]), nodef, leaf)
		}
	}
}
