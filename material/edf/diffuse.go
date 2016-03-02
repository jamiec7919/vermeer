package edf

import (
	"github.com/jamiec7919/vermeer/material"
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

type Diffuse struct {
	E [3]float32
}

func (b *Diffuse) Eval(surf *material.SurfacePoint, omega_o m.Vec3, Le *material.Spectrum) error {
	d := omega_o[2]
	Le.FromRGB(b.E[0]*d, b.E[1]*d, b.E[2]*d)
	return nil
}

func (b *Diffuse) Sample(surf *material.SurfacePoint, rnd *rand.Rand, omega_o *m.Vec3, Le *material.Spectrum, pdf *float64) error {
	return nil
}
