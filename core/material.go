package core

import (
	"github.com/jamiec7919/vermeer/colour"
	m "github.com/jamiec7919/vermeer/math"
)

type MaterialId int32

type Material interface {
	Name() string
	SetId(id int32)
	Id() int32
	//ApplyBumpMap(surf *SurfacePoint)
	HasBumpMap() bool

	Emission(sg *ShaderGlobals, omega_o m.Vec3) colour.RGB

	Eval(sg *ShaderGlobals)
}
