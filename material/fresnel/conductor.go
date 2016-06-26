// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fresnel

import (
	"github.com/jamiec7919/vermeer/colour"
	m "github.com/jamiec7919/vermeer/math"
)

/*
Conductor is a Fresnel model for metals/conductors.

Algorithm from 'Artist Friendly Metallic Fresnel' by Ole Gulbrandsen.

http://jcgt.org/published/0003/04/03/paper.pdf
*/
type Conductor struct {
	r, g colour.RGB
}

// NewConductor returns a new conductor model.
//
// model is ignored for now, may be used in future for presets. See paper
// for meaning of r and g, essentially are reflectance and edge tint colours.
func NewConductor(model int, r, g colour.RGB) *Conductor {
	return &Conductor{r, g}
}

func nMin(r float32) float32 {
	return (1 - r) / (1 + r)
}

func nMax(r float32) float32 {
	return (1 + m.Sqrt(r)) / (1 - m.Sqrt(r))
}

func getN(r, g float32) float32 {
	return nMin(r)*g + (1-g)*nMax(r)
}

func getK2(r, n float32) float32 {
	nr := (n+1)*(n+1)*r - (n-1)*(n-1)
	return nr / (1 - r)
}

func getR(n, k float32) float32 {
	return ((n-1)*(n-1) + k*k) / ((n+1)*(n+1) + k*k)
}

func getG(n, k float32) float32 {
	r := getR(n, k)

	return (nMax(r) - n) / (nMax(r) - nMin(r))
}

func fresnel(r, g, cosTheta float32) float32 {
	nr := m.Clamp(r, 0, 0.99)

	n := getN(nr, g)
	k2 := getK2(nr, n)

	rsNum := n*n + k2 - 2*n*cosTheta + cosTheta*cosTheta
	rsDen := n*n + k2 + 2*n*cosTheta + cosTheta*cosTheta

	rs := rsNum / rsDen

	rpNum := (n*n+k2)*cosTheta*cosTheta - 2*n*cosTheta + 1
	rpDen := (n*n+k2)*cosTheta*cosTheta + 2*n*cosTheta + 1
	rp := rpNum / rpDen

	return 0.5 * (rs + rp)
}

// Kr is the Fresnel coefficient for this material.
//
// Implements core.Fresnel
//
// cosTheta is the clamped dot product of direction and surface normal.
//
// Note that this returns an RGB value.  To get a single value use RGB.Maxh().
func (f *Conductor) Kr(cosTheta float32) (out colour.RGB) {
	for k := range out {
		out[k] = fresnel(f.r[k], f.g[k], cosTheta)
	}

	return
}
