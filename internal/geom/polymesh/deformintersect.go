// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package polymesh

import (
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/qbvh"
)

///go:nosplit
func (mesh *PolyMesh) intersectRayMotion(ray *core.RayData, sg *core.ShaderGlobals) int32 {

	k := ray.Time * float32(mesh.Verts.MotionKeys-1)

	time := k - m.Floor(k)

	key := int(m.Floor(k))
	key2 := int(m.Ceil(k))
	// if key == key2 then no interpolation needed but probably still do it to avoid branch
	//	log.Printf("%v %v %v %v %v", k, m.Ceil(k), time, key, key2)
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
			pnode := &(mesh.accel.mqbvh.Nodes[node])

			for i := range ray.Supp.Boxes {
				ray.Supp.Boxes[i] = (1.0-time)*mesh.accel.mqbvh.Boxes[key][node][i] + time*mesh.accel.mqbvh.Boxes[key2][node][i]
			}

			intersectBoxes(&ray.Ray, &ray.Supp.Boxes, &ray.Supp.Hits, &ray.Supp.T)

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

				faceidx := int(mesh.accel.idx[i])

				face.V[0] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+0])+(mesh.Verts.ElemsPerKey*key)],
					mesh.Verts.Elems[int(mesh.idxp[faceidx*3+0])+(mesh.Verts.ElemsPerKey*key2)], time)

				face.V[1] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+1])+(mesh.Verts.ElemsPerKey*key)],
					mesh.Verts.Elems[int(mesh.idxp[faceidx*3+1])+(mesh.Verts.ElemsPerKey*key2)], time)

				face.V[2] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+2])+(mesh.Verts.ElemsPerKey*key)],
					mesh.Verts.Elems[int(mesh.idxp[faceidx*3+2])+(mesh.Verts.ElemsPerKey*key2)], time)

				face.PrimID = uint64(faceidx)

				//log.Printf("%v", face)
				if face.IntersectRay(ray) {
					ray.Result.MtlID = mesh.mtlid
					hit = true

					if mesh.UV.Elems != nil {
						k := ray.Time * float32(mesh.UV.MotionKeys-1)

						time := k - m.Floor(k)

						key := int(m.Floor(k))
						key2 := int(m.Ceil(k))

						for j := range face.UV {
							face.UV[j] = m.Vec2Lerp(mesh.UV.Elems[int(mesh.uvtriidx[faceidx*3+j])+(mesh.UV.ElemsPerKey*key)],
								mesh.UV.Elems[int(mesh.uvtriidx[faceidx*3+j])+(mesh.UV.ElemsPerKey*key2)], time)
						}
					} else {
						face.UV[0] = m.Vec2{0, 0}
						face.UV[1] = m.Vec2{1, 0}
						face.UV[2] = m.Vec2{0, 1}
					}

					if mesh.Normals.Elems != nil {
						k := ray.Time * float32(mesh.Normals.MotionKeys-1)

						time := k - m.Floor(k)

						key := int(m.Floor(k))
						key2 := int(m.Ceil(k))

						for j := range face.Ns {
							face.Ns[j] = m.Vec3Lerp(mesh.Normals.Elems[int(mesh.normalidx[faceidx*3+j])+(mesh.Normals.ElemsPerKey*key)],
								mesh.
									Normals.Elems[int(mesh.normalidx[faceidx*3+j])+(mesh.Normals.ElemsPerKey*key2)], time)
						}
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
func (mesh *PolyMesh) intersectRayEpsilonMotion(ray *core.RayData, sg *core.ShaderGlobals, epsilon float32) int32 {

	k := ray.Time * float32(mesh.Verts.MotionKeys-1)

	time := k - m.Floor(k)

	key := int(m.Floor(k))
	key2 := int(m.Ceil(k))
	// if key == key2 then no interpolation needed but probably still do it to avoid branch

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
			pnode := &(mesh.accel.mqbvh.Nodes[node])

			for i := range ray.Supp.Boxes {
				ray.Supp.Boxes[i] = (1.0-time)*mesh.accel.mqbvh.Boxes[key][node][i] + time*mesh.accel.mqbvh.Boxes[key2][node][i]
			}

			intersectBoxes(&ray.Ray, &ray.Supp.Boxes, &ray.Supp.Hits, &ray.Supp.T)

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

				faceidx := int(mesh.accel.idx[i])

				face.V[0] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+0])+(mesh.Verts.ElemsPerKey*key)],
					mesh.Verts.Elems[int(mesh.idxp[faceidx*3+0])+(mesh.Verts.ElemsPerKey*key2)], time)

				face.V[1] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+1])+(mesh.Verts.ElemsPerKey*key)],
					mesh.Verts.Elems[int(mesh.idxp[faceidx*3+1])+(mesh.Verts.ElemsPerKey*key2)], time)

				face.V[2] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+2])+(mesh.Verts.ElemsPerKey*key)],
					mesh.Verts.Elems[int(mesh.idxp[faceidx*3+2])+(mesh.Verts.ElemsPerKey*key2)], time)

				face.PrimID = uint64(faceidx)

				if face.IntersectRayEpsilon(ray, epsilon) {
					ray.Result.MtlID = mesh.mtlid
					hit = true
					k := ray.Time * float32(mesh.UV.MotionKeys-1)

					time := k - m.Floor(k)

					key := int(m.Floor(k))
					key2 := int(m.Ceil(k))

					for j := range face.UV {
						face.UV[j] = m.Vec2Lerp(mesh.UV.Elems[int(mesh.uvtriidx[faceidx*3+j])+(mesh.UV.ElemsPerKey*key)],
							mesh.UV.Elems[int(mesh.uvtriidx[faceidx*3+j])+(mesh.UV.ElemsPerKey*key2)], time)
					}
				} else {
					face.UV[0] = m.Vec2{0, 0}
					face.UV[1] = m.Vec2{1, 0}
					face.UV[2] = m.Vec2{0, 1}
				}

				if mesh.Normals.Elems != nil {
					k := ray.Time * float32(mesh.Normals.MotionKeys-1)

					time := k - m.Floor(k)

					key := int(m.Floor(k))
					key2 := int(m.Ceil(k))

					for j := range face.Ns {
						face.Ns[j] = m.Vec3Lerp(mesh.Normals.Elems[int(mesh.normalidx[faceidx*3+j])+(mesh.Normals.ElemsPerKey*key)],
							mesh.
								Normals.Elems[int(mesh.normalidx[faceidx*3+j])+(mesh.Normals.ElemsPerKey*key2)], time)
					}
				} else {
					face.Ns[0] = face.N
					face.Ns[1] = face.N
					face.Ns[2] = face.N
				}

				face.shaderParams(ray, sg)
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

//go:nosplit
func (mesh *PolyMesh) intersectVisRayMotion(ray *core.RayData) bool {

	k := ray.Time * float32(mesh.Verts.MotionKeys-1)

	time := k - m.Floor(k)

	key := int(m.Floor(k))
	key2 := int(m.Ceil(k))
	// if key == key2 then no interpolation needed but probably still do it to avoid branch

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
			pnode := &(mesh.accel.mqbvh.Nodes[node])

			for i := range ray.Supp.Boxes {
				ray.Supp.Boxes[i] = (1.0-time)*mesh.accel.mqbvh.Boxes[key][node][i] + time*mesh.accel.mqbvh.Boxes[key2][node][i]
			}

			intersectBoxes(&ray.Ray, &ray.Supp.Boxes, &ray.Supp.Hits, &ray.Supp.T)

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

				faceidx := int(mesh.accel.idx[i])

				face.V[0] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+0])+(mesh.Verts.ElemsPerKey*key)],
					mesh.Verts.Elems[int(mesh.idxp[faceidx*3+0])+(mesh.Verts.ElemsPerKey*key2)], time)

				face.V[1] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+1])+(mesh.Verts.ElemsPerKey*key)],
					mesh.Verts.Elems[int(mesh.idxp[faceidx*3+1])+(mesh.Verts.ElemsPerKey*key2)], time)

				face.V[2] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+2])+(mesh.Verts.ElemsPerKey*key)],
					mesh.Verts.Elems[int(mesh.idxp[faceidx*3+2])+(mesh.Verts.ElemsPerKey*key2)], time)

				if face.IntersectVisRay(ray) {
					ray.Ray.Tclosest = 0.5

					return true
				}
			}

		}
	}
	return false
}

//go:nosplit
func (mesh *PolyMesh) intersectVisRayEpsilonMotion(ray *core.RayData, epsilon float32) bool {

	k := ray.Time * float32(mesh.Verts.MotionKeys-1)

	time := k - m.Floor(k)

	key := int(m.Floor(k))
	key2 := int(m.Ceil(k))
	// if key == key2 then no interpolation needed but probably still do it to avoid branch

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
			pnode := &(mesh.accel.mqbvh.Nodes[node])

			for i := range ray.Supp.Boxes {
				ray.Supp.Boxes[i] = (1.0-time)*mesh.accel.mqbvh.Boxes[key][node][i] + time*mesh.accel.mqbvh.Boxes[key2][node][i]
			}

			intersectBoxes(&ray.Ray, &ray.Supp.Boxes, &ray.Supp.Hits, &ray.Supp.T)

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

				faceidx := int(mesh.accel.idx[i])

				face.V[0] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+0])+(mesh.Verts.ElemsPerKey*key)],
					mesh.Verts.Elems[int(mesh.idxp[faceidx*3+0])+(mesh.Verts.ElemsPerKey*key2)], time)

				face.V[1] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+1])+(mesh.Verts.ElemsPerKey*key)],
					mesh.Verts.Elems[int(mesh.idxp[faceidx*3+1])+(mesh.Verts.ElemsPerKey*key2)], time)

				face.V[2] = m.Vec3Lerp(mesh.Verts.Elems[int(mesh.idxp[faceidx*3+2])+(mesh.Verts.ElemsPerKey*key)],
					mesh.Verts.Elems[int(mesh.idxp[faceidx*3+2])+(mesh.Verts.ElemsPerKey*key2)], time)

				if face.IntersectVisRayEpsilon(ray, epsilon) {
					ray.Ray.Tclosest = 0.5

					return true
				}
			}

		}
	}
	return false
}
