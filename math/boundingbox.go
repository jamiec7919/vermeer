// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// BoundingBox represents a 3D axis-aligned bounding box.
type BoundingBox struct {
	Bounds [2][3]float32
}

// InfBox is a null sized box at +ve infinty.  This box is impossible to hit.
// This is provided as cannot use the Reset box (+inf,-inf) with the optimized box/ray intersection.
var InfBox = BoundingBox{Bounds: [2][3]float32{[3]float32{Inf(1), Inf(1), Inf(1)}, [3]float32{Inf(1), Inf(1), Inf(1)}}}

// SurfaceArea returns the surface area of the box.
func (b *BoundingBox) SurfaceArea() float32 {
	return b.Dim(0)*b.Dim(1)*2.0 + b.Dim(1)*b.Dim(2)*2 + b.Dim(0)*b.Dim(2)*2
}

// Dim returns the size of the box in the given dimension (0=x,1=y,2=z).
func (b *BoundingBox) Dim(axis int) float32 {
	return b.Bounds[1][axis] - b.Bounds[0][axis]
}

// Centroid returns the 3D point at the centre of the box.
func (b *BoundingBox) Centroid() (o Vec3) {
	o[0] = (b.Bounds[1][0] + b.Bounds[0][0]) * 0.5
	o[1] = (b.Bounds[1][1] + b.Bounds[0][1]) * 0.5
	o[2] = (b.Bounds[1][2] + b.Bounds[0][2]) * 0.5
	return
}

// AxisCentroid returns the centroid of the box projected on the axis.
func (b *BoundingBox) AxisCentroid(axis int) float32 {
	return (b.Bounds[1][axis] + b.Bounds[0][axis]) * 0.5
}

// Centre returns the centroid of the box projected on the axis.
//
// Deprecated: Centre is the same as Centroid.
func (b *BoundingBox) Centre(axis int) float32 {
	return b.AxisCentroid(axis)
}

// MaxDim returns the axis in which the box is maximum.
func (b *BoundingBox) MaxDim() int {
	return b.LongestAxis()
}

// LongestAxis returns the axis in which the box is maximum.
//
// Deprecated: LongestAxis is the same as MaxDim.
func (b *BoundingBox) LongestAxis() (axis int) {

	if b.Dim(0) < b.Dim(1) {
		if b.Dim(1) < b.Dim(2) {
			axis = 2
		} else {
			axis = 1
		}
	} else {
		if b.Dim(0) < b.Dim(2) {
			axis = 2
		} else {
			axis = 0
		}
	}
	return
}

// ResetDim will reset the box in the given dimension only.
func (b *BoundingBox) ResetDim(dim int) {
	b.Bounds[0][dim] = Inf(1)
	b.Bounds[1][dim] = Inf(-1)

}

// GrowDim will expand the box to contain the point projected on the given axis.
func (b *BoundingBox) GrowDim(dim int, x float32) {
	b.Bounds[0][dim] = Min(b.Bounds[0][dim], x)
	b.Bounds[1][dim] = Max(b.Bounds[1][dim], x)
}

// Reset will set the box to an invalid infinite size that bounds no volume.
func (b *BoundingBox) Reset() {
	for i := 0; i < 3; i++ {
		b.Bounds[0][i] = Inf(1)
		b.Bounds[1][i] = Inf(-1)
	}
}

// GrowVec3 will expand the box to contain the given point represented as a Vec3.
func (b *BoundingBox) GrowVec3(P Vec3) {
	b.Grow(P[0], P[1], P[2])
}

// GrowBox will grow the box to contain the given BoundingBox.
func (b *BoundingBox) GrowBox(p BoundingBox) {
	for k := 0; k < 3; k++ {
		b.Bounds[0][k] = Min(b.Bounds[0][k], p.Bounds[0][k])
		b.Bounds[1][k] = Max(b.Bounds[1][k], p.Bounds[1][k])
	}
}

// Grow will expand the box to contain the given point.
func (b *BoundingBox) Grow(X, Y, Z float32) {
	b.Bounds[0][0] = Min(X, b.Bounds[0][0])
	b.Bounds[1][0] = Max(X, b.Bounds[1][0])

	b.Bounds[0][1] = Min(Y, b.Bounds[0][1])
	b.Bounds[1][1] = Max(Y, b.Bounds[1][1])
	b.Bounds[0][2] = Min(Z, b.Bounds[0][2])
	b.Bounds[1][2] = Max(Z, b.Bounds[1][2])

}
