package core

import (
	"github.com/jamiec7919/vermeer/colour"
	m "github.com/jamiec7919/vermeer/math"
)

// MaterialId is the type of shader ids.
type MaterialId int32

// Material represents a surface shader (Note: this will be renamed to Shader or SurfaceShader).
type Material interface {
	Name() string
	SetId(id int32)
	Id() int32
	//ApplyBumpMap(surf *SurfacePoint)
	HasBumpMap() bool

	Emission(sg *ShaderGlobals, omega_o m.Vec3) colour.RGB

	// Eval evaluates the shader and returns values in sh.OutXXX members.
	Eval(sg *ShaderGlobals)
}
