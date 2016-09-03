// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qbvh

import (
	"github.com/jamiec7919/vermeer/core"
)

//go:nosplit
//go:noescape
func intersectBoxes(ray *core.Ray, boxes *[4 * 3 * 2]float32, hits *[4]int32, t *[4]float32)

// Trace traverses the tree and calls prim.TraceElems for any non-empty leaf encountered.
//go:nosplit
func Trace(qbvh []Node, prim Primitive, ray *core.Ray, sg *core.ShaderContext) bool {
	// Push root node on stack:
	stackTop := ray.Task.Traversal.StackTop
	ray.Task.Traversal.Stack[stackTop].Node = 0
	ray.Task.Traversal.Stack[stackTop].T = ray.Tclosest

	hit := false

	for stackTop >= ray.Task.Traversal.StackTop {

		node := ray.Task.Traversal.Stack[stackTop].Node
		T := ray.Task.Traversal.Stack[stackTop].T
		stackTop--

		if ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(qbvh[node])

			intersectBoxes(ray, &pnode.Boxes, &ray.Task.Traversal.Hits, &ray.Task.Traversal.T)

			//order := [4]int{0, 1, 2, 3} // actually in reverse order as this is order pushed on stack
			var order [4]uint8

			if ray.D[pnode.Axis0] < 0 {
				if ray.D[pnode.Axis2] < 0 {
					order[3] = 3
					order[2] = 2
				} else {
					order[3] = 2
					order[2] = 3
				}
				if ray.D[pnode.Axis1] < 0 {
					order[1] = 1
					order[0] = 0
				} else {
					order[1] = 0
					order[0] = 1
				}
			} else {
				if ray.D[pnode.Axis2] < 0 {
					order[1] = 3
					order[0] = 2
				} else {
					order[1] = 2
					order[0] = 3
				}
				if ray.D[pnode.Axis1] < 0 {
					order[3] = 1
					order[2] = 0
				} else {
					order[3] = 0
					order[2] = 1
				}

			}

			//_ = ray.Task.Traversal.Stack[stackTop+4]

			for j := range order {
				k := order[j]
				if ray.Task.Traversal.Hits[k] != 0 {
					stackTop++
					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[k]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[k]

				}

			}

		} else if node < -1 {
			// Leaf
			leafBase := LeafBase(node)
			leafCount := LeafCount(node)

			// Save the stacktop in case primitive wants to reuse stack
			tmp := ray.Task.Traversal.StackTop
			ray.Task.Traversal.StackTop = stackTop + 1 // at this point stackTop points to the next valid node on stack so need +1

			if prim.TraceElems(ray, sg, leafBase, leafCount) {
				hit = true

				// Early out if this ray is a shadow ray
				if ray.Type&core.RayTypeShadow != 0 {
					// restore
					ray.Task.Traversal.StackTop = tmp
					return true
				}
			}

			// restore
			ray.Task.Traversal.StackTop = tmp
		}
	}

	return hit

}
