// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package material

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

type EDF interface {
	Eval(*core.SurfacePoint, m.Vec3, *colour.Spectrum) error
	Sample(*core.SurfacePoint, *rand.Rand, *m.Vec3, *colour.Spectrum, *float64) error // Sample direction given point
}
