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

const MaxLeafCount = 16

type Index int32 // Type of a triangle/prim index

// 128 byte struct
type Node struct {
	Boxes               [4 * 3 * 2]float32
	Axis0, Axis1, Axis2 int32
	Children            [4]int32
	Parent              int32
}

func (n *Node) SetBounds(idx int, bounds m.BoundingBox) {
	//log.Printf("SetBounds %v %v", idx, bounds)
	for i := 0; i < 2; i++ {
		for k := 0; k < 3; k++ {
			n.Boxes[idx+(i*12)+(k*4)] = bounds.Bounds[i][k]
		}
	}
	//log.Printf("SetBounds %v %v", idx, n.boxes)
}

func (n *Node) Bounds(idx int) (bounds m.BoundingBox) {
	for i := 0; i < 2; i++ {
		for k := 0; k < 3; k++ {
			bounds.Bounds[i][k] = n.Boxes[idx+(i*12)+(k*4)]
		}
	}

	return
}
func LEAF_COUNT(l int32) int { return int((((l) & 0xf) + 1)) }

func LEAF_BASE(l int32) int { return int(((l) & 0x7ffffff) >> 4) }

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

// count <= 16
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

func (n *Node) SetChild(idx int, ch int32) {
	n.Children[idx] = ch
}

func (n *Node) SetAxis0(axis int32) {
	n.Axis0 = axis
}

func (n *Node) SetAxis1(axis int32) {
	n.Axis1 = axis
}

func (n *Node) SetAxis2(axis int32) {
	n.Axis2 = axis
}

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
