// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qbvh

import (
	m "github.com/jamiec7919/vermeer/math"
)

/*
Need algorithm to clip arbitrary convex polygon (starting from triangle) to axis aligned plane.

This is silly, should split both sides at once since we know we're definitely splitting the primitive and
want both sides!!!!
*/

// ClipLeft clips the polygon represented by verts by the axis-aligned plane described by d and axis.
// Keep the bits to the left of the plane.
func ClipLeft(d float32, axis int, verts []m.Vec3) (outverts []m.Vec3) {
	outverts = make([]m.Vec3, 0, len(verts))

	for i := range verts {
		inext := (i + 1) % len(verts)

		vd0 := verts[i][axis] - d
		vd1 := verts[inext][axis] - d

		if vd0 > 0 && vd1 > 0 { // don't output either vertex, just move on

		} else if vd0 <= 0 && vd1 <= 0 { // both verts are fine, output them
			outverts = append(outverts, verts[i]) // Note: inext will be output on next iteration
		} else if vd0 == 0 && vd1 > 0 {

			// vert[i] lies on plane, just output it
			outverts = append(outverts, verts[i])

		} else if vd0 < 0 && vd1 > 0 {
			// verts[i] is fine but clip the edge and output clipped vertex but not verts[inext]
			// need the t value where edge[axis] = d
			t := -vd0 / (verts[inext][axis] - verts[i][axis])

			vnew := m.Vec3Mad(verts[i], m.Vec3Sub(verts[inext], verts[i]), t)
			outverts = append(outverts, verts[i])
			outverts = append(outverts, vnew)

		} else if vd0 > 0 && vd1 < 0 {
			// verts[inext] is fine but clip the edge and output clipped vertex but not verts[i]
			// need the t value where edge[axis] = d
			t := -vd0 / (verts[inext][axis] - verts[i][axis])

			vnew := m.Vec3Mad(verts[i], m.Vec3Sub(verts[inext], verts[i]), t)
			outverts = append(outverts, vnew)
			//outverts = append(outverts, verts[inext])
		}

	}
	return
}

// ClipRight clips the polygon represented by verts by the axis-aligned plane described by d and axis.
// Keep the bits to the right of the plane.
func ClipRight(d float32, axis int, verts []m.Vec3) (outverts []m.Vec3) {
	outverts = make([]m.Vec3, 0, len(verts))

	for i := range verts {
		inext := (i + 1) % len(verts)

		vd0 := verts[i][axis] - d
		vd1 := verts[inext][axis] - d

		if vd0 < 0 && vd1 < 0 { // don't output either vertex, just move on

		} else if vd0 >= 0 && vd1 >= 0 { // both verts are fine, output them
			outverts = append(outverts, verts[i]) // Note: inext will be output on next iteration
		} else if vd0 == 0 && vd1 < 0 {

			// vert[i] lies on plane, just output it
			outverts = append(outverts, verts[i])

		} else if vd0 > 0 && vd1 < 0 {
			// verts[i] is fine but clip the edge and output clipped vertex but not verts[inext]
			// need the t value where edge[axis] = d
			t := -vd0 / (verts[inext][axis] - verts[i][axis])

			vnew := m.Vec3Mad(verts[i], m.Vec3Sub(verts[inext], verts[i]), t)
			outverts = append(outverts, verts[i])
			outverts = append(outverts, vnew)

		} else if vd0 < 0 && vd1 > 0 {
			// verts[inext] is fine but clip the edge and output clipped vertex but not verts[i]
			// need the t value where edge[axis] = d
			t := -vd0 / (verts[inext][axis] - verts[i][axis])

			vnew := m.Vec3Mad(verts[i], m.Vec3Sub(verts[inext], verts[i]), t)
			outverts = append(outverts, vnew)
			//outverts = append(outverts, verts[inext])
		}

	}
	return
}
