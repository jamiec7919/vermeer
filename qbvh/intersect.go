// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qbvh

import (
	//"fmt"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
)

//go:nosplit
//go:noescape
func intersectBoxes(ray *core.Ray, boxes *[4 * 3 * 2]float32, hits *[4]int32, t *[4]float32)

func intersectBoxesSlow(ray *core.Ray, boxes *[4 * 3 * 2]float32, hits *[4]int32, t *[4]float32) {
	var sign [3]uint8

	for k := 0; k < 3; k++ {
		if ray.Dinv.Elt(k) < 0.0 {
			sign[k] = 1
		}
	}

	for idx := uint8(0); idx < 4; idx++ {

		// x's are: boxes[idx+(sign[0]*12)+0]
		// y's are: boxes[idx+(sign[1]*12)+4]
		// z's are: boxes[idx+(sign[2]*12)+8]
		tmin := (boxes[idx+(sign[0]*12)+0] - ray.P.X) * ray.Dinv.X
		tmax := (boxes[idx+((1-sign[0])*12)+0] - ray.P.X) * ray.Dinv.X
		tymin := (boxes[idx+(sign[1]*12)+4] - ray.P.Y) * ray.Dinv.Y
		tymax := (boxes[idx+((1-sign[1])*12)+4] - ray.P.Y) * ray.Dinv.Y
		tzmin := (boxes[idx+(sign[2]*12)+8] - ray.P.Z) * ray.Dinv.Z
		tzmax := (boxes[idx+((1-sign[2])*12)+8] - ray.P.Z) * ray.Dinv.Z

		tNear := m.Max(m.Max(tmin, tymin), m.Max(0.0, tzmin))
		tFar := m.Min(m.Min(tmax, tymax), m.Min(ray.Tclosest, tzmax))

		(*t)[idx] = tNear
		if tNear <= tFar {
			(*hits)[idx] = -1
		} else {
			(*hits)[idx] = 0

		}
	}

}

func intersectBoxesSlow2(ray *core.Ray, boxes *[4 * 3 * 2]float32, hits *[4]int32, t *[4]float32) {

	for idx := uint8(0); idx < 4; idx++ {

		// x's are: boxes[idx+(sign[0]*12)+0]
		// y's are: boxes[idx+(sign.Y*12)+4]
		// z's are: boxes[idx+(sign.Z*12)+8]
		tx1 := (boxes[idx+(0*12)+0] - ray.P.X) * ray.Dinv.X
		tx2 := (boxes[idx+(1*12)+0] - ray.P.X) * ray.Dinv.X

		tmin := m.Min(tx1, tx2)
		tmax := m.Max(tx1, tx2)

		ty1 := (boxes[idx+(0*12)+4] - ray.P.Y) * ray.Dinv.Y
		ty2 := (boxes[idx+(1*12)+4] - ray.P.Y) * ray.Dinv.Y

		tmin = m.Max(tmin, m.Min(ty1, ty2))
		tmax = m.Min(tmax, m.Max(ty1, ty2))

		tz1 := (boxes[idx+(0*12)+8] - ray.P.Z) * ray.Dinv.Z
		tz2 := (boxes[idx+(1*12)+8] - ray.P.Z) * ray.Dinv.Z

		tmin = m.Max(tmin, m.Min(tz1, tz2))
		tmax = m.Min(tmax, m.Max(tz1, tz2))

		(*t)[idx] = m.Max(0, tmin)

		if tmax >= m.Max(0, tmin) {
			(*hits)[idx] = -1
			//			(*t)[idx] = tmin
		} else {
			(*hits)[idx] = 0
		}
	}

}

// Trace traverses the tree and calls prim.TraceElems for any non-empty leaf encountered.
//go:nosplit
func Trace(qbvh []Node, prim Primitive, ray *core.Ray, sg *core.ShaderContext) bool {
	// Push root node on stack:
	stackTop := ray.Task.Traversal.StackTop
	ray.Task.Traversal.Stack[stackTop].Node = 0
	ray.Task.Traversal.Stack[stackTop].T = ray.Tclosest

	hit := false

	stackTop++

	for stackTop > ray.Task.Traversal.StackTop {

		stackTop--
		node := ray.Task.Traversal.Stack[stackTop].Node

		if ray.Tclosest < ray.Task.Traversal.Stack[stackTop].T || node == -1 {
			//stackTop-- // pop the top, it isn't interesting
			//node = -1 // pretend we're an empty leaf
			continue
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			ray.NodesT++
			intersectBoxes(ray, &qbvh[node].Boxes, &ray.Task.Traversal.Hits, &ray.Task.Traversal.T)
			/*
				if false {
					var hits, hits2 [4]int32
					var t, t2 [4]float32
					intersectBoxesSlow(ray, &qbvh[node].Boxes, &hits, &t)
					intersectBoxesSlow2(ray, &qbvh[node].Boxes, &hits2, &t2)

					ok := true
					for k := range hits {
						if hits[k] != ray.Task.Traversal.Hits[k] {
							fmt.Printf("*%v %v %v %v %v %v %v %v\n", ray.Task.Traversal.T, t, ray.Task.Traversal.Hits, hits, ray.P, ray.D, qbvh[node].Children)
							ok = false
						}
					}
					if ok {
						fmt.Printf("%v %v %v %v %v %v %v\n", ray.Task.Traversal.T, t, ray.Task.Traversal.Hits, hits, ray.P, ray.D, qbvh[node].Children)
					}
				}*/
			//order := [4]int{0, 1, 2, 3} // actually in reverse order as this is order pushed on stack
			//var order [4]uint8

			if ray.D.Elt(int(qbvh[node].Axis0)) < 0 {
				if ray.D.Elt(int(qbvh[node].Axis1)) < 0 {

					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[0]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[0]
					stackTop -= ray.Task.Traversal.Hits[0]

					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[1]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[1]
					stackTop -= ray.Task.Traversal.Hits[1]

				} else {
					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[1]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[1]
					stackTop -= ray.Task.Traversal.Hits[1]

					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[0]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[0]
					stackTop -= ray.Task.Traversal.Hits[0]
				}

				if ray.D.Elt(int(qbvh[node].Axis2)) < 0 {
					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[2]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[2]
					stackTop -= ray.Task.Traversal.Hits[2]

					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[3]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[3]
					stackTop -= ray.Task.Traversal.Hits[3]
				} else {
					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[3]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[3]
					stackTop -= ray.Task.Traversal.Hits[3]

					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[2]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[2]
					stackTop -= ray.Task.Traversal.Hits[2]

				}
			} else {
				if ray.D.Elt(int(qbvh[node].Axis2)) < 0 {

					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[2]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[2]
					stackTop -= ray.Task.Traversal.Hits[2]

					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[3]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[3]
					stackTop -= ray.Task.Traversal.Hits[3]
				} else {
					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[3]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[3]
					stackTop -= ray.Task.Traversal.Hits[3]

					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[2]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[2]
					stackTop -= ray.Task.Traversal.Hits[2]

				}
				if ray.D.Elt(int(qbvh[node].Axis1)) < 0 {
					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[0]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[0]
					stackTop -= ray.Task.Traversal.Hits[0]

					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[1]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[1]
					stackTop -= ray.Task.Traversal.Hits[1]

				} else {
					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[1]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[1]
					stackTop -= ray.Task.Traversal.Hits[1]

					ray.Task.Traversal.Stack[stackTop].Node = qbvh[node].Children[0]
					ray.Task.Traversal.Stack[stackTop].T = ray.Task.Traversal.T[0]
					stackTop -= ray.Task.Traversal.Hits[0]

				}

			}

		} else /*if node < -1*/ {
			// Leaf
			//leafBase := LeafBase(node)
			//leafCount := LeafCount(node)
			ray.LeafsT++

			// Save the stacktop in case primitive wants to reuse stack
			tmp := ray.Task.Traversal.StackTop
			ray.Task.Traversal.StackTop = stackTop + 1 // at this point stackTop points to the next valid node on stack so need +1

			if prim.TraceElems(ray, sg, LeafBase(node), LeafCount(node)) {
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
