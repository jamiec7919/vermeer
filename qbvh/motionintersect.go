// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qbvh

import (
	"github.com/jamiec7919/vermeer/core"
)

/*
	k := ray.Time * float32(mesh.Verts.MotionKeys-1)

	time := k - m.Floor(k)

	key := int(m.Floor(k))
	key2 := int(m.Ceil(k))
*/

// TraceMotion intersects a MotionQBVH. time is [0,1) keys relate to boxes[].
//go:nosplit
func TraceMotion(qbvh MotionQBVH, time float32, key, key2 int, prim MotionPrimitive, ray *core.Ray, sg *core.ShaderContext) bool {

	// if key == key2 then no interpolation needed but probably still do it to avoid branch
	//	log.Printf("%v %v %v %v %v", k, m.Ceil(k), time, key, key2)
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
			pnode := &(qbvh.Nodes[node])

			for i := range ray.Task.Traversal.Boxes {
				ray.Task.Traversal.Boxes[i] = (1.0-time)*qbvh.Boxes[key][node][i] + time*qbvh.Boxes[key2][node][i]
			}

			intersectBoxes(ray, &ray.Task.Traversal.Boxes, &ray.Task.Traversal.Hits, &ray.Task.Traversal.T)

			order := [4]int{0, 1, 2, 3} // actually in reverse order as this is order pushed on stack

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

			for j := range order {
				k := order[j]
				if ray.Task.Traversal.Hits[k] != 0 {
					stackTop++
					ray.Task.Traversal.Stack[stackTop].Node = pnode.Children[k]
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

			if prim.TraceMotionElems(time, key, key2, ray, sg, leafBase, leafCount) {
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
