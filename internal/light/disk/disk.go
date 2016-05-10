// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package light

import (
	"errors"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/internal/geom/mesh"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/nodes"
)

type Disk struct {
	NodeName      string `node:"Name"`
	P, Up, LookAt m.Vec3
	T, B, N       m.Vec3
	Radius        float32
	Material      string
	MtlId         int32
}

var ErrNoSample = errors.New("No smaple")

func (d *Disk) DiffuseShadeMult() float32 {
	return 1.0
}
func (d *Disk) Name() string { return d.NodeName }
func (d *Disk) PreRender(rc *core.RenderContext) error {
	mtlid := rc.GetMaterialId(d.Material)

	if mtlid == -1 {
		return errors.New("DiskLight: can't find material " + d.Material)
	}
	d.MtlId = mtlid

	mesh := d.CreateDisk(rc, d.P, d.LookAt, d.Up, d.Radius, mtlid)
	rc.AddNode(mesh)
	return nil
}
func (d *Disk) PostRender(rc *core.RenderContext) error { return nil }

/*
func (d *Disk) SamplePoint(rnd *rand.Rand, surf *core.SurfacePoint, pdf *float64) error {
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
*/
func (d *Disk) SampleArea(sg *core.ShaderGlobals) error {
	r0 := sg.Rand().Float32()
	r1 := sg.Rand().Float32()

	u := d.Radius * m.Sqrt(r0) * m.Cos(2*m.Pi*r1)
	v := d.Radius * m.Sqrt(r0) * m.Sin(2*m.Pi*r1)

	pdf := float64(1.0 / (m.Pi * d.Radius * d.Radius))

	P := m.Vec3Add3(d.P, m.Vec3Scale(u, d.B), m.Vec3Scale(v, d.T))

	V := m.Vec3Sub(P, sg.P)

	if m.Vec3Dot(V, sg.Ng) > 0.0 && m.Vec3Dot(V, d.N) < 0.0 {
		sg.Ldist = m.Vec3Length(V)
		sg.Ld = m.Vec3Normalize(V)

		lightm := core.GetMaterial(d.MtlId)

		sg.Liu.Lambda = sg.Lambda
		//lightm.EvalEDF(&P, P.WorldToTangent(m.Vec3Neg(sg.Ld)), &sg.Liu)
		omega_o := m.Vec3BasisProject(d.B, d.T, d.N, m.Vec3Neg(sg.Ld))
		dot_n := omega_o[2]
		E := lightm.Emission(sg, omega_o)
		sg.Liu.FromRGB(E[0]*dot_n, E[1]*dot_n, E[2]*dot_n)

		// geometry term / pdf
		sg.Weight = m.Abs(m.Vec3Dot(sg.Ld, sg.N)) * m.Abs(m.Vec3Dot(sg.Ld, d.N)) / (sg.Ldist * sg.Ldist)
		sg.Weight /= float32(pdf)

		return nil
	} else {
		return ErrNoSample

	}
}

/*
func (d *Disk) SampleDirection(surf *core.SurfacePoint, rnd *rand.Rand, omega_o *m.Vec3, Le *colour.Spectrum, pdf *float64) error {
	return nil
}
*/
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

func (d *Disk) CreateDisk(rc *core.RenderContext, P, t, up m.Vec3, radius float32, mtlid int32) *mesh.StaticMesh {
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

	d.N = N
	d.T = T
	d.B = B
	return &mesh.StaticMesh{"lightmesh", &msh}

}

func init() {
	nodes.Register("DiskLight", func() (core.Node, error) {

		return &Disk{}, nil

	})
}
