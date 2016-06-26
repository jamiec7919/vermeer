// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package polymesh

import (
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/qbvh"
)

//var _ = core.Primitive(&PolyMesh{})

//go:nosplit
//go:noescape
func intersectBoxes(ray *core.Ray, boxes *[4 * 3 * 2]float32, hits *[4]int32, t *[4]float32)

// TraceRay implements core.Primitive.
func (mesh *PolyMesh) TraceRay(ray *core.RayData, sg *core.ShaderGlobals) int32 {

	if mesh.RayBias > 0.0 {
		if mesh.Verts.MotionKeys > 1 {
			return mesh.intersectRayEpsilonMotion(ray, sg, mesh.RayBias)
		}
		return mesh.intersectRayEpsilon(ray, sg, mesh.RayBias)

	}

	if mesh.Verts.MotionKeys > 1 {
		return mesh.intersectRayMotion(ray, sg)
	}

	return mesh.intersectRay(ray, sg)

}

// VisRay implements core.Primitive.
func (mesh *PolyMesh) VisRay(ray *core.RayData) {
	if mesh.RayBias > 0.0 {
		if mesh.Verts.MotionKeys > 1 {
			mesh.intersectVisRayEpsilonMotion(ray, mesh.RayBias)
		} else {
			mesh.intersectVisRayEpsilon(ray, mesh.RayBias)
		}
	} else {
		if mesh.Verts.MotionKeys > 1 {
			mesh.intersectVisRayMotion(ray)
		} else {
			mesh.intersectVisRay(ray)
		}
	}

}

///go:nosplit
func (mesh *PolyMesh) intersectRay(ray *core.RayData, sg *core.ShaderGlobals) int32 {
	// Push root node on stack:
	stackTop := 0
	ray.Supp.Stack[stackTop].Node = 0
	ray.Supp.Stack[stackTop].T = ray.Ray.Tclosest

	hit := false

	for stackTop >= 0 {

		node := ray.Supp.Stack[stackTop].Node
		T := ray.Supp.Stack[stackTop].T
		stackTop--

		if ray.Ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(mesh.accel.qbvh[node])

			intersectBoxes(&ray.Ray, &pnode.Boxes, &ray.Supp.Hits, &ray.Supp.T)

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
					ray.Supp.Stack[stackTop].Node = pnode.Children[k]
					ray.Supp.Stack[stackTop].T = ray.Supp.T[k]

				}

			}

		} else if node < -1 {
			// Leaf
			leafBase := qbvh.LeafBase(node)
			leafCount := qbvh.LeafCount(node)

			for i := leafBase; i < leafBase+leafCount; i++ {
				var face Face

				faceidx := mesh.accel.idx[i]

				face.V[0] = mesh.Verts.Elems[mesh.idxp[faceidx*3+0]]
				face.V[1] = mesh.Verts.Elems[mesh.idxp[faceidx*3+1]]
				face.V[2] = mesh.Verts.Elems[mesh.idxp[faceidx*3+2]]
				face.PrimID = uint64(faceidx)

				if face.IntersectRay(ray) {
					hit = true

					if mesh.UV.Elems != nil {
						face.UV[0] = mesh.UV.Elems[mesh.uvtriidx[faceidx*3+0]]
						face.UV[1] = mesh.UV.Elems[mesh.uvtriidx[faceidx*3+1]]
						face.UV[2] = mesh.UV.Elems[mesh.uvtriidx[faceidx*3+2]]
					} else {
						face.UV[0] = m.Vec2{0, 0}
						face.UV[1] = m.Vec2{1, 0}
						face.UV[2] = m.Vec2{0, 1}
					}
					if mesh.Normals.Elems != nil {
						face.Ns[0] = mesh.Normals.Elems[mesh.normalidx[faceidx*3+0]]
						face.Ns[1] = mesh.Normals.Elems[mesh.normalidx[faceidx*3+1]]
						face.Ns[2] = mesh.Normals.Elems[mesh.normalidx[faceidx*3+2]]
					} else {
						face.Ns[0] = face.N
						face.Ns[1] = face.N
						face.Ns[2] = face.N
					}

					face.shaderParams(ray, sg)
				}
			}

		}
	}
	if hit {
		sg.Poffset = ray.Result.POffset
		sg.P = ray.Result.P
		sg.N = ray.Result.Ns
		sg.Ns = ray.Result.Ns

		sg.Ng = ray.Result.Ng
		//sg.U = ray.Result.UV[0]
		//sg.V = ray.Result.UV[1]
		//sg.DdPdu = ray.Result.Pu
		//sg.DdPdv = ray.Result.Pv
		return ray.Result.MtlID
	}
	return -1

}

///go:nosplit
func (mesh *PolyMesh) intersectRayEpsilon(ray *core.RayData, sg *core.ShaderGlobals, epsilon float32) int32 {
	// Push root node on stack:
	stackTop := 0
	ray.Supp.Stack[stackTop].Node = 0
	ray.Supp.Stack[stackTop].T = ray.Ray.Tclosest

	hit := false

	for stackTop >= 0 {

		node := ray.Supp.Stack[stackTop].Node
		T := ray.Supp.Stack[stackTop].T
		stackTop--

		if ray.Ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(mesh.accel.qbvh[node])

			intersectBoxes(&ray.Ray, &pnode.Boxes, &ray.Supp.Hits, &ray.Supp.T)

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
					ray.Supp.Stack[stackTop].Node = pnode.Children[k]
					ray.Supp.Stack[stackTop].T = ray.Supp.T[k]

				}

			}

		} else if node < -1 {
			// Leaf
			leafBase := qbvh.LeafBase(node)
			leafCount := qbvh.LeafCount(node)

			for i := leafBase; i < leafBase+leafCount; i++ {
				var face Face

				faceidx := mesh.accel.idx[i]

				face.V[0] = mesh.Verts.Elems[mesh.idxp[faceidx*3+0]]
				face.V[1] = mesh.Verts.Elems[mesh.idxp[faceidx*3+1]]
				face.V[2] = mesh.Verts.Elems[mesh.idxp[faceidx*3+2]]
				face.PrimID = uint64(faceidx)

				if face.IntersectRayEpsilon(ray, epsilon) {
					hit = true

					if mesh.UV.Elems != nil {
						face.UV[0] = mesh.UV.Elems[mesh.uvtriidx[faceidx*3+0]]
						face.UV[1] = mesh.UV.Elems[mesh.uvtriidx[faceidx*3+1]]
						face.UV[2] = mesh.UV.Elems[mesh.uvtriidx[faceidx*3+2]]
					} else {
						face.UV[0] = m.Vec2{0, 0}
						face.UV[1] = m.Vec2{1, 0}
						face.UV[2] = m.Vec2{0, 1}
					}

					if mesh.Normals.Elems != nil {
						face.Ns[0] = mesh.Normals.Elems[mesh.normalidx[faceidx*3+0]]
						face.Ns[1] = mesh.Normals.Elems[mesh.normalidx[faceidx*3+1]]
						face.Ns[2] = mesh.Normals.Elems[mesh.normalidx[faceidx*3+2]]
					} else {
						face.Ns[0] = face.N
						face.Ns[1] = face.N
						face.Ns[2] = face.N
					}

					face.shaderParams(ray, sg)
				}
			}

		}
	}
	if hit {
		sg.Poffset = ray.Result.POffset
		sg.P = ray.Result.P
		sg.N = ray.Result.Ns
		sg.Ns = ray.Result.Ns

		sg.Ng = ray.Result.Ng
		//	sg.U = ray.Result.UV[0]
		//	sg.V = ray.Result.UV[1]
		//	sg.DdPdu = ray.Result.Pu
		//	sg.DdPdv = ray.Result.Pv
		return ray.Result.MtlID
	}
	return -1

}

//go:nosplit
func (mesh *PolyMesh) intersectVisRay(ray *core.RayData) bool {
	// Push root node on stack:
	stackTop := 0
	ray.Supp.Stack[stackTop].Node = 0
	ray.Supp.Stack[stackTop].T = ray.Ray.Tclosest

	for stackTop >= 0 {

		node := ray.Supp.Stack[stackTop].Node
		T := ray.Supp.Stack[stackTop].T
		stackTop--

		if ray.Ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(mesh.accel.qbvh[node])

			intersectBoxes(&ray.Ray, &pnode.Boxes, &ray.Supp.Hits, &ray.Supp.T)

			for k := range pnode.Children {
				if ray.Supp.Hits[k] != 0 {
					stackTop++
					ray.Supp.Stack[stackTop].Node = pnode.Children[k]
					ray.Supp.Stack[stackTop].T = ray.Supp.T[k]

				}

			}

		} else if node < -1 {
			// Leaf
			leafBase := qbvh.LeafBase(node)
			leafCount := qbvh.LeafCount(node)

			for i := leafBase; i < leafBase+leafCount; i++ {
				var face Face
				faceidx := mesh.accel.idx[i]

				face.V[0] = mesh.Verts.Elems[mesh.idxp[faceidx*3+0]]
				face.V[1] = mesh.Verts.Elems[mesh.idxp[faceidx*3+1]]
				face.V[2] = mesh.Verts.Elems[mesh.idxp[faceidx*3+2]]

				if face.IntersectVisRay(ray) {
					return true
				}
			}

		}
	}
	return false
}

//go:nosplit
func (mesh *PolyMesh) intersectVisRayEpsilon(ray *core.RayData, epsilon float32) bool {
	// Push root node on stack:
	stackTop := 0
	ray.Supp.Stack[stackTop].Node = 0
	ray.Supp.Stack[stackTop].T = ray.Ray.Tclosest

	for stackTop >= 0 {

		node := ray.Supp.Stack[stackTop].Node
		T := ray.Supp.Stack[stackTop].T
		stackTop--

		if ray.Ray.Tclosest < T {
			//stackTop-- // pop the top, it isn't interesting
			node = -1 // pretend we're an empty leaf
		}
		// We already know ray intersects this node, so check all children and push onto stack if ray intersects.

		if node >= 0 {
			pnode := &(mesh.accel.qbvh[node])

			intersectBoxes(&ray.Ray, &pnode.Boxes, &ray.Supp.Hits, &ray.Supp.T)

			for k := range pnode.Children {

				if ray.Supp.Hits[k] != 0 {
					stackTop++
					ray.Supp.Stack[stackTop].Node = pnode.Children[k]
					ray.Supp.Stack[stackTop].T = ray.Supp.T[k]

				}

			}

		} else if node < -1 {
			// Leaf
			leafBase := qbvh.LeafBase(node)
			leafCount := qbvh.LeafCount(node)

			for i := leafBase; i < leafBase+leafCount; i++ {
				var face Face
				faceidx := mesh.accel.idx[i]

				face.V[0] = mesh.Verts.Elems[mesh.idxp[faceidx*3+0]]
				face.V[1] = mesh.Verts.Elems[mesh.idxp[faceidx*3+1]]
				face.V[2] = mesh.Verts.Elems[mesh.idxp[faceidx*3+2]]

				if face.IntersectVisRayEpsilon(ray, epsilon) {
					return true
				}
			}

		}
	}
	return false
}
