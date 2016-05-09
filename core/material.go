package core

import (
	"github.com/jamiec7919/vermeer/colour"
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

type MaterialId int32

type Material interface {
	Name() string
	SetId(id int32)
	Id() int32
	ApplyBumpMap(surf *SurfacePoint)
	HasEDF() bool
	HasBumpMap() bool
	IsDelta(surf *SurfacePoint) bool
	EvalEDF(surf *SurfacePoint, omega_o m.Vec3, Le *colour.Spectrum) error
	EvalBSDF(surf *SurfacePoint, omega_i, omega_o m.Vec3, rho *colour.Spectrum) error
	SampleBSDF(surf *SurfacePoint, omega_i m.Vec3, rnd *rand.Rand, omega_o *m.Vec3, rho *colour.Spectrum, pdf *float64) error
}
