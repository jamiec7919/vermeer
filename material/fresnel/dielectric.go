// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fresnel

import (
	"github.com/jamiec7919/vermeer/colour"
	m "github.com/jamiec7919/vermeer/math"
)

func sqr32(v float32) float32 { return v * v }

// Dielectric is a Fresnel model for common dielectric materials.
type Dielectric struct {
	eta float32 // eta = n_t/n_i, e.g. n_t = 1.5 for glass, n_i = 1 for air.

}

// NewDielectric returns a new dielectric model.
//
// eta is the index-of-refraction ratio (n_t/n_i).
func NewDielectric(eta float32) *Dielectric {
	return &Dielectric{eta}
}

// Kr for the given direction and normal.
//
// Implements core.Fresnel.
//
// cosTheta is the clamped dot product of direction and surface normal.
//
// Note that this returns an RGB value.  To get a single value use RGB.Maxh().
func (f *Dielectric) Kr(cosTheta float32) colour.RGB {
	c := cosTheta
	g := (f.eta * f.eta) - 1 + (c * c)

	if g < 0.0 {
		return colour.RGB{1, 1, 1}
	}

	g = m.Sqrt(g)

	fr := (1.0 / 2.0) * sqr32((g-c)/(g+c)) * (1 + sqr32((c*(g+c)-1)/(c*(g-c)+1)))

	return colour.RGB{fr, fr, fr}
}
