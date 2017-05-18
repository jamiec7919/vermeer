// Copyright 2017 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import (
	"unsafe"
)

// Elt returns the i'th component of the vector.
func (v *Vec3) Elt(i int) float32 {
	return (*(*[3]float32)(unsafe.Pointer(v)))[i]

}

// Set returns the i'th component of the vector.
func (v *Vec3) Set(i int, a float32) {
	(*(*[3]float32)(unsafe.Pointer(v)))[i] = a

}
