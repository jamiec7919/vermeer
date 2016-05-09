// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edf

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

type Diffuse struct {
	E [3]float32
}

func (b *Diffuse) Eval(surf *core.SurfacePoint, omega_o m.Vec3, Le *colour.Spectrum) error {
	d := omega_o[2]
	Le.FromRGB(b.E[0]*d, b.E[1]*d, b.E[2]*d)
	return nil
}

func (b *Diffuse) Sample(surf *core.SurfacePoint, rnd *rand.Rand, omega_o *m.Vec3, Le *colour.Spectrum, pdf *float64) error {
	return nil
}
