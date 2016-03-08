// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package light

import (
	"github.com/jamiec7919/vermeer/internal/core"
	"github.com/jamiec7919/vermeer/internal/geom/mesh"
	"github.com/jamiec7919/vermeer/material"
	"github.com/jamiec7919/vermeer/material/bsdf"
	"github.com/jamiec7919/vermeer/material/edf"
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

type Disk struct {
	NodeName string
	P        m.Vec3
	T, B, N  m.Vec3
	Radius   float32
	EDF      material.EDF
	MtlId    material.Id
}

func (d *Disk) Name() string                            { return d.NodeName }
func (d *Disk) PreRender(rc *core.RenderContext) error  { return nil }
func (d *Disk) PostRender(rc *core.RenderContext) error { return nil }

func (d *Disk) SamplePoint(rnd *rand.Rand, surf *material.SurfacePoint, pdf *float64) error {
	r0 := rnd.Float32()
	r1 := rnd.Float32()

	u := d.Radius * m.Sqrt(r0) * m.Cos(2*m.Pi*r1)
	v := d.Radius * m.Sqrt(r0) * m.Sin(2*m.Pi*r1)

	*pdf = float64(1.0 / (m.Pi * d.Radius * d.Radius))

	P := m.Vec3Add3(d.P, m.Vec3Scale(u, d.B), m.Vec3Scale(v, d.T))

	surf.P = P
	surf.N = d.N
	surf.B = d.B
	surf.T = d.T
	surf.Ns = d.N
	surf.MtlId = int32(d.MtlId)

	return nil
}
func (d *Disk) SampleArea(from *material.SurfacePoint, rnd *rand.Rand, surf *material.SurfacePoint, pdf *float64) error {
	r0 := rnd.Float32()
	r1 := rnd.Float32()

	u := d.Radius * m.Sqrt(r0) * m.Cos(2*m.Pi*r1)
	v := d.Radius * m.Sqrt(r0) * m.Sin(2*m.Pi*r1)

	*pdf = float64(1.0 / (m.Pi * d.Radius * d.Radius))

	P := m.Vec3Add3(d.P, m.Vec3Scale(u, d.B), m.Vec3Scale(v, d.T))

	surf.P = P
	surf.N = d.N
	surf.B = d.B
	surf.T = d.T
	surf.Ns = d.N
	surf.MtlId = int32(d.MtlId)

	return nil
}

func (d *Disk) SampleDirection(surf *material.SurfacePoint, rnd *rand.Rand, omega_o *m.Vec3, Le *material.Spectrum, pdf *float64) error {
	if d.EDF == nil {
		return core.ErrNoSample
	}

	return d.EDF.Sample(surf, rnd, omega_o, Le, pdf)
}

/*
// shade is in space of point to need to project
func (d *Disk) Sample(shade *core.ShadePoint, rnd *rand.Rand, sample *core.DirectionalSample) error {

	r0 := rnd.Float32()
	r1 := rnd.Float32()

	u := d.Radius * m.Sqrt(r0) * m.Cos(2*m.Pi*r1)
	v := d.Radius * m.Sqrt(r0) * m.Sin(2*m.Pi*r1)

	sample.PDF = float64(1.0 / (m.Pi * d.Radius * d.Radius))

	x := m.Vec3Add3(d.P, m.Vec3Scale(u, d.B), m.Vec3Scale(v, d.T))

	V := m.Vec3Normalize(m.Vec3Sub(x, shade.O))

	if m.Vec3Dot(V, shade.Ng) < 0 || m.Vec3Dot(V, d.N) > -0.1 {
		return core.ErrNoSample
	}
	sample.Omega = m.Vec3BasisProject(shade.Tg, shade.Bg, shade.Ng, V)

	// Omega is Area PDF, need to project to projected solid angle
	// p_perp * cos theta_o = p_A * cos theta_i / || x - x' ||^2
	cos_thetai := m.Abs(m.Vec3Dot(d.N, V))
	//cos_omega := m.Abs(sample.Omega[2])
	len2 := m.Vec3Length2(m.Vec3Sub(x, shade.O))

	sample.PDF = float64((float32(sample.PDF) * cos_thetai) / len2)
	return nil
}
*/
func (d *Disk) Pos() m.Vec3 {
	return d.P
}

func CreateDisk(rc *core.RenderContext, P, t, up m.Vec3, radius float32, mtlid material.Id) interface{} {
	N := m.Vec3Normalize(m.Vec3Sub(t, P))
	T := m.Vec3Normalize(m.Vec3Cross(N, up))
	B := m.Vec3Cross(N, T)

	nv := 10

	dv := 2 * m.Pi / float32(nv)

	msh := mesh.Mesh{}
	ang := float32(0)

	for i := 0; i < nv; i++ {
		face := mesh.FaceGeom{}
		face.V[0] = P
		face.V[2] = m.Vec3Add(P, m.Vec3Add(m.Vec3Scale(radius*m.Cos(ang), B), m.Vec3Scale(radius*m.Sin(ang), T)))
		face.V[1] = m.Vec3Add(P, m.Vec3Add(m.Vec3Scale(radius*m.Cos(ang+dv), B), m.Vec3Scale(radius*m.Sin(ang+dv), T)))
		face.MtlId = int32(mtlid)
		ang += dv
		//face.setup()
		//log.Printf("%v", face)
		msh.Faces = append(msh.Faces, face)
	}

	//mesh.InitAccel()
	/*
		prim := core.Primitive{
			Mesh:        &mesh,
			M:           m.MatrixIdentity,
			Minv:        m.MatrixIdentity,
			WorldBounds: mesh.WorldBounds(m.MatrixIdentity),
		}
	*/
	node := &Disk{P: P, T: T, N: N, B: B, Radius: radius, MtlId: mtlid}

	return []core.Node{node, &mesh.StaticMesh{"lightmesh", &msh}}

}

func init() {
	core.RegisterType("DiskLight", func(rc *core.RenderContext, params core.Params) (interface{}, error) {

		diff2 := bsdf.Diffuse{Kd: &material.ConstantMap{[3]float32{0.8, 0.8, 0.8}}}
		edf := edf.Diffuse{E: [3]float32{100, 100, 100}}
		mtl2 := material.Material{Sides: 1, BSDF: [2]material.BSDF{&diff2}, EDF: &edf}
		mtlid := rc.AddMaterial("light1", &mtl2)

		P, _ := params.GetVec3("P")
		LookAt, _ := params.GetVec3("LookAt")
		Up, _ := params.GetVec3("Up")
		radius, _ := params.GetFloat("Radius")
		return CreateDisk(rc, P, LookAt, Up, float32(radius), mtlid), nil

	})
}
