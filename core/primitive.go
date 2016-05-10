// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/qbvh"
)

type Primitive interface {
	TraceRay(*RayData, *ShaderGlobals) int32
	VisRay(*RayData)
	WorldBounds() m.BoundingBox
	Visible() bool

	// UVCoord returns the UV coordinate for the given set, given the element id and surface params.
	// UVCoord(set int, elem uint32, su,sv float32) m.Vec3
}

//go:nosplit
//go:noescape
func rayNodeIntersectAll_asm(ray *Ray, node *qbvh.Node, hit *[4]int32, tNear *[4]float32)

//go:nosplit
func (scene *Scene) traceRayAccel(ray *RayData, sg *ShaderGlobals) (mtlid int32) {
	// Push root node on stack:
	stackTop := 0
	ray.Supp.TopLevelStack[stackTop].Node = 0
	ray.Supp.TopLevelStack[stackTop].T = ray.Ray.Tclosest
	mtlid = -1

	for stackTop >= 0 {

		node := ray.Supp.TopLevelStack[stackTop].Node
		T := ray.Supp.TopLevelStack[stackTop].T
		stackTop--

		if ray.Ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(scene.nodes[node])
			rayNodeIntersectAll_asm(&ray.Ray, pnode, &ray.Supp.Hits, &ray.Supp.T)

			order := [4]int{0, 1, 2, 3} // actually in reverse order as this is order pushed on stack

			if ray.Ray.D[pnode.Axis0] < 0 {
				if ray.Ray.D[pnode.Axis2] < 0 {
					order[3] = 3
					order[2] = 2
				} else {
					order[3] = 2
					order[2] = 3
				}
				if ray.Ray.D[pnode.Axis1] < 0 {
					order[1] = 1
					order[0] = 0
				} else {
					order[1] = 0
					order[0] = 1
				}
			} else {
				if ray.Ray.D[pnode.Axis2] < 0 {
					order[1] = 3
					order[0] = 2
				} else {
					order[1] = 2
					order[0] = 3
				}
				if ray.Ray.D[pnode.Axis1] < 0 {
					order[3] = 1
					order[2] = 0
				} else {
					order[3] = 0
					order[2] = 1
				}

			}

			for j := range order {
				k := order[j]

				if ray.Supp.Hits[k] != 0 {
					stackTop++
					ray.Supp.TopLevelStack[stackTop].Node = pnode.Children[k]
					ray.Supp.TopLevelStack[stackTop].T = ray.Supp.T[k]

				} else {
					//log.Printf("Miss %v %v", node, pnode.Children[k])
				}

			}

		} else if node < -1 {
			// Leaf
			leaf_base := qbvh.LEAF_BASE(node)
			leaf_count := qbvh.LEAF_COUNT(node)
			// log.Printf("leaf %v,%v: %v %v", traverseStack[stackTop].node, k, leaf_base, leaf_count)
			for i := leaf_base; i < leaf_base+leaf_count; i++ {
				_mtlid := scene.prims[i].TraceRay(ray, sg)

				if _mtlid > -1 {
					ray.Result.Prim = scene.prims[i]
					mtlid = _mtlid
				}
			}
		}
	}
	return
}

//go:nosplit
func (scene *Scene) visRayAccel(ray *RayData) {
	// Push root node on stack:
	stackTop := 0
	ray.Supp.TopLevelStack[stackTop].Node = 0
	ray.Supp.TopLevelStack[stackTop].T = ray.Ray.Tclosest

	for stackTop >= 0 {

		node := ray.Supp.TopLevelStack[stackTop].Node
		T := ray.Supp.TopLevelStack[stackTop].T
		stackTop--

		if ray.Ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(scene.nodes[node])
			rayNodeIntersectAll_asm(&ray.Ray, pnode, &ray.Supp.Hits, &ray.Supp.T)

			for k := range pnode.Children {
				if ray.Supp.Hits[k] != 0 {
					stackTop++
					ray.Supp.TopLevelStack[stackTop].Node = pnode.Children[k]
					ray.Supp.TopLevelStack[stackTop].T = ray.Supp.T[k]
				}

			}

		} else if node < -1 {
			// Leaf
			leaf_base := qbvh.LEAF_BASE(node)
			leaf_count := qbvh.LEAF_COUNT(node)
			// log.Printf("leaf %v,%v: %v %v", traverseStack[stackTop].node, k, leaf_base, leaf_count)
			for i := leaf_base; i < leaf_base+leaf_count; i++ {
				scene.prims[i].VisRay(ray)

				if !ray.IsVis() {
					return
				}
			}
		}
	}

}
