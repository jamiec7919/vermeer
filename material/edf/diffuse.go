// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package edf provides the built-in EDFs (Emission Distribution Functions) for Vermeer.
*/
package edf

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

// Diffuse implements basic diffuse distribution (cosine weighted).
type Diffuse struct {
	E [3]float32
}

// Eval evaluates the emission in the given direction.
func (b *Diffuse) Eval(surf *core.SurfacePoint, omegaO m.Vec3, Le *colour.Spectrum) error {
	d := omegaO[2]
	Le.FromRGB(b.E[0]*d, b.E[1]*d, b.E[2]*d)
	return nil
}

// Sample samples the emission distribution.
//
// Deprecated: unused.
func (b *Diffuse) Sample(surf *core.SurfacePoint, rnd *rand.Rand, omegaO *m.Vec3, Le *colour.Spectrum, pdf *float64) error {
	return nil
}
