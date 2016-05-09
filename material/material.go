// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package material

import (
	"errors"
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/material/edf"
	"github.com/jamiec7919/vermeer/material/texture"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/nodes"
	"math/rand"
)

type Id int32

const ID_NONE Id = -1

type Material struct {
	MtlName string `node:"Name"`
	id      int32  // This should only be assinged by RenderContext

	Sides             int
	Specular          string
	Diffuse           string
	Ks, Kd            core.MapSampler
	Roughness         core.MapSampler
	SpecularRoughness core.MapSampler
	IOR               core.MapSampler
	E                 core.MapSampler

	BSDF [2]BSDF // bsdf for sides
	//Medium [2]Medium  // medium material
	EDF EDF // Emission distribution for light materials (nil is not a light)

	BumpMap *BumpMap
}

// core.Node methods
func (m *Material) Name() string { return m.MtlName }
func (m *Material) PreRender(rc *core.RenderContext) error {
	bsdf := makeBSDF(m)

	if bsdf == nil {
		return errors.New("BSDF " + m.Diffuse + " " + m.Specular + " not found")
	}
	m.BSDF[0] = bsdf

	if m.E != nil {
		edf := &edf.Diffuse{E: m.E.SampleRGB(0, 0, 0, 0)}
		m.EDF = edf
	}
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
	return m.EDF != nil
}

func (m *Material) HasBumpMap() bool {
	return m.BumpMap != nil
}

func (m *Material) IsDelta(surf *core.SurfacePoint) bool {
	return m.BSDF[0].IsDelta(surf)
}

func (m *Material) EvalEDF(surf *core.SurfacePoint, omega_o m.Vec3, Le *colour.Spectrum) error {
	return m.EDF.Eval(surf, omega_o, Le)
}

func (m *Material) EvalBSDF(surf *core.SurfacePoint, omega_i, omega_o m.Vec3, rho *colour.Spectrum) error {
	return m.BSDF[0].Eval(surf, omega_i, omega_o, rho)
}

func (m *Material) SampleBSDF(surf *core.SurfacePoint, omega_i m.Vec3, rnd *rand.Rand, omega_o *m.Vec3, rho *colour.Spectrum, pdf *float64) error {
	return m.BSDF[0].Sample(surf, omega_i, rnd, omega_o, rho, pdf)
}

type BumpMap struct {
	Map   core.MapSampler
	Scale float32
}

func (mtl *Material) ApplyBumpMap(surf *core.SurfacePoint) {
	delta := float32(1) / float32(6000)
	//log.Printf("Bump %v %v %v %v", mtl.BumpMap.Map.SampleRGB(surf.UV[0][0]-delta, surf.UV[0][1], delta, delta)[0], mtl.BumpMap.Map.SampleRGB(surf.UV[0][0]+delta, surf.UV[0][1], delta, delta)[0], surf.UV[0][0]-delta, surf.UV[0][0]+delta)
	Bu := (1.0 / (2.0 * delta)) * mtl.BumpMap.Scale * (mtl.BumpMap.Map.SampleRGB(surf.UV[0][0]-delta, surf.UV[0][1], delta, delta)[0] - mtl.BumpMap.Map.SampleRGB(surf.UV[0][0]+delta, surf.UV[0][1], delta, delta)[0])
	Bv := (1.0 / (2.0 * delta)) * mtl.BumpMap.Scale * (mtl.BumpMap.Map.SampleRGB(surf.UV[0][0], surf.UV[0][1]-delta, delta, delta)[0] - mtl.BumpMap.Map.SampleRGB(surf.UV[0][0], surf.UV[0][1]+delta, delta, delta)[0])
	//log.Printf("Bump %v %v %v", Bu, Bv, surf.Ns)
	surf.Ns = m.Vec3Add(surf.Ns, m.Vec3Sub(m.Vec3Scale(Bu, m.Vec3Cross(surf.Ns, surf.Pv[0])), m.Vec3Scale(Bv, m.Vec3Cross(surf.Ns, surf.Pu[0]))))
	//log.Printf("%v", surf.Ns)
	surf.SetupTangentSpace(surf.Ns)

}

type ConstantMap struct {
	C [3]float32
}

func (c *ConstantMap) SampleRGB(s, t, ds, dt float32) (out [3]float32) {
	out[0] = c.C[0]
	out[1] = c.C[1]
	out[2] = c.C[2]
	return
}

func (c *ConstantMap) SampleScalar(s, t, ds, dt float32) (out float32) {
	out = c.C[0]
	return
}

type TextureMap struct {
	Filename string
}

func (c *TextureMap) SampleRGB(s, t, ds, dt float32) (out [3]float32) {
	return texture.SampleRGB(c.Filename, s, t, ds, dt)
}

func (c *TextureMap) SampleScalar(s, t, ds, dt float32) (out float32) {
	q := texture.SampleRGB(c.Filename, s, t, ds, dt)
	return q[0]

}

func makeMaterial() (core.Node, error) {

	mtl := &Material{}
	return mtl, nil
}

func init() {
	nodes.Register("Material", makeMaterial)
}
