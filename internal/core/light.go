// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/jamiec7919/vermeer/material"
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

type Light interface {
	SamplePoint(*rand.Rand, *material.SurfacePoint, *float64) error                                  // Sample a point on the surface
	SampleArea(*material.SurfacePoint, *rand.Rand, *material.SurfacePoint, *float64) error           // Sample a point on the surface visible from first point
	SampleDirection(*material.SurfacePoint, *rand.Rand, *m.Vec3, *material.Spectrum, *float64) error // Sample direction given point

}
