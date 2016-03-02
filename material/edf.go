// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package material

import (
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

type EDF interface {
	Eval(*SurfacePoint, m.Vec3, *Spectrum) error
	Sample(*SurfacePoint, *rand.Rand, *m.Vec3, *Spectrum, *float64) error // Sample direction given point
}
