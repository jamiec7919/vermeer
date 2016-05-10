// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package material

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/nodes"
	//"log"
)

type Id int32

const ID_NONE Id = -1

type Material struct {
	MtlName string `node:"Name"`
	id      int32  // This should only be assinged by RenderContext

	Sides             int
	Specular          string
	Diffuse           string
	Ks, Kd            core.RGBParam     // should be RGBParam
	Roughness         core.Float32Param // .. FloatParam
	SpecularRoughness core.Float32Param
	IOR               core.Float32Param
	E                 core.RGBParam

	//Medium [2]Medium  // medium material
	BumpMapScale float32
	BumpMap      core.Float32Param

	//	BumpMap *BumpMap
}

// core.Node methods
func (m *Material) Name() string { return m.MtlName }
func (m *Material) PreRender(rc *core.RenderContext) error {
	return nil
}
func (m *Material) PostRender(rc *core.RenderContext) error { return nil }

// core.Material methods
func (m *Material) Id() int32 {
	return m.id
}

func (m *Material) SetId(id int32) {
	m.id = id
}

func (m *Material) HasEDF() bool {
	return m.E != nil
}

func (m *Material) HasBumpMap() bool {
	return m.BumpMap != nil
}

/*
func (m *Material) EvalEDF(surf *core.SurfacePoint, omega_o m.Vec3, Le *colour.Spectrum) error {
	return nil
}

func (m *Material) EvalBSDF(surf *core.SurfacePoint, omega_i, omega_o m.Vec3, rho *colour.Spectrum) error {
	return nil
}

func (m *Material) SampleBSDF(surf *core.SurfacePoint, omega_i m.Vec3, rnd *rand.Rand, omega_o *m.Vec3, rho *colour.Spectrum, pdf *float64) error {
	return nil
}
*/
func (m *Material) Emission(sg *core.ShaderGlobals, omega_o m.Vec3) colour.RGB {
	return m.E.RGB(sg)
}

type BumpMap struct {
	Map   core.Float32Param
	Scale float32
}

func (mtl *Material) ApplyBumpMap(sg *core.ShaderGlobals) {
	u := sg.U
	v := sg.V

	delta := float32(1) / float32(6000)

	sg.V -= delta
	tv0 := mtl.BumpMap.Float32(sg)
	sg.V = v + delta
	tv1 := mtl.BumpMap.Float32(sg)
	sg.V = v
	sg.U -= delta
	tu0 := mtl.BumpMap.Float32(sg)
	sg.U = u + delta
	tu1 := mtl.BumpMap.Float32(sg)
	sg.U = u

	//log.Printf("Bump %v %v %v %v", mtl.BumpMap.Map.SampleRGB(surf.UV[0][0]-delta, surf.UV[0][1], delta, delta)[0], mtl.BumpMap.Map.SampleRGB(surf.UV[0][0]+delta, surf.UV[0][1], delta, delta)[0], surf.UV[0][0]-delta, surf.UV[0][0]+delta)
	Bu := (1.0 / (2.0 * delta)) * mtl.BumpMapScale * (tu0 - tu1)
	Bv := (1.0 / (2.0 * delta)) * mtl.BumpMapScale * (tv0 - tv1)
	//log.Printf("Bump %v %v %v", Bu, Bv, surf.Ns)
	//Q := sg.N
	V := m.Vec3Normalize(m.Vec3Cross(sg.N, sg.DdPdu))
	U := m.Vec3Cross(sg.N, V)

	sg.N = m.Vec3Add(sg.N, m.Vec3Sub(m.Vec3Scale(Bu, m.Vec3Cross(sg.N, U)), m.Vec3Scale(Bv, m.Vec3Cross(sg.N, V))))
	sg.N = m.Vec3Normalize(sg.N)
	//log.Printf("%v %v", sg.N, Q)
	//sg.SetupTangentSpace(sg.N)

}

func makeMaterial() (core.Node, error) {

	mtl := &Material{}
	return mtl, nil
}

func init() {
	nodes.Register("Material", makeMaterial)
}
