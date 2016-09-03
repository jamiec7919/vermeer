// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package light

import (
	"fmt"
	"github.com/jamiec7919/vermeer/builtin/geom/polymesh"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/math/ldseq"
	"github.com/jamiec7919/vermeer/nodes"
)

// Disk represents a circular disk light node.
type Disk struct {
	NodeDef       core.NodeDef
	NodeName      string `node:"Name"`
	P, Up, LookAt m.Vec3
	T, B, N       m.Vec3
	Radius        float32
	Shader        string
	Segments      int
	Samples       int

	shader core.Shader
	geom   core.Geom
}

var _ core.Node = (*Disk)(nil)
var _ core.Light = (*Disk)(nil)

// Name implements core.Node.
func (d *Disk) Name() string { return d.NodeName }

// Def implements core.Node.
func (d *Disk) Def() core.NodeDef { return d.NodeDef }

// PreRender implelments core.Node.
func (d *Disk) PreRender() error {

	if s := core.FindNode(d.Shader); s != nil {
		shader, ok := s.(core.Shader)

		if !ok {
			return fmt.Errorf("Unable to find shader %v", d.Shader)
		}

		d.shader = shader

		d.N = m.Vec3Normalize(m.Vec3Sub(d.LookAt, d.P))
		d.T = m.Vec3Normalize(m.Vec3Cross(d.N, d.Up))
		d.B = m.Vec3Cross(d.N, d.T)

		mesh := d.createMesh()
		core.AddNode(mesh)
		d.geom = mesh

	} else {
		return fmt.Errorf("Unable to find node (shader %v)", d.Shader)

	}

	return nil
}

// PotentialContrib implements core.Light.
func (d *Disk) PotentialContrib(sg *core.ShaderContext) float32 {

	return 0.5
}

// NumSamples implements core.Light
func (d *Disk) NumSamples(sg *core.ShaderContext) int {
	return d.Samples * d.Samples
}

// Geom implements core.Light
func (d *Disk) Geom() core.Geom { return d.geom }

// PostRender implelments core.Node.
func (d *Disk) PostRender() error { return nil }

// SampleArea implements core.Light.
func (d *Disk) SampleArea(sg *core.ShaderContext, n int) error {

	idx := uint64(sg.I*sg.NSamples + sg.Sample)
	r0 := ldseq.VanDerCorput(idx, sg.SampleScramble)
	r1 := ldseq.Sobol(idx, sg.SampleScramble2)

	u := d.Radius * m.Sqrt(float32(r0)) * m.Cos(2*m.Pi*float32(r1))
	v := d.Radius * m.Sqrt(float32(r0)) * m.Sin(2*m.Pi*float32(r1))

	pdf := float64(1.0 / (m.Pi * d.Radius * d.Radius))

	P := m.Vec3Add3(d.P, m.Vec3Scale(u, d.B), m.Vec3Scale(v, d.T))

	V := m.Vec3Sub(P, sg.P)

	if m.Vec3Dot(V, sg.Ng) > 0.0 && m.Vec3Dot(V, d.N) < 0.0 {
		sg.Ldist = m.Vec3Length(V)
		sg.Ld = m.Vec3Normalize(V)

		sg.Liu.Lambda = sg.Lambda
		omegaO := m.Vec3BasisProject(d.B, d.T, d.N, m.Vec3Neg(sg.Ld))
		//ODotN := omegaO[2]
		E := d.shader.EvalEmission(sg, omegaO)
		//		sg.Liu.FromRGB(E[0]*ODotN, E[1]*ODotN, E[2]*ODotN)
		//E.Scale(m.Abs(omegaO[2]))
		sg.Liu.FromRGB(E)

		// geometry term / pdf
		sg.Weight = m.Abs(m.Vec3Dot(sg.Ld, sg.N)) * m.Abs(m.Vec3Dot(sg.Ld, d.N)) / (sg.Ldist * sg.Ldist)
		sg.Weight /= float32(pdf)

		//		fmt.Printf("%v %v %v %v\n", sg.Poffset, sg.P, pdf, sg.Weight)
		return nil
	}

	return fmt.Errorf("nosample")
}

// DiffuseShadeMult implements core.Light.
func (d *Disk) DiffuseShadeMult() float32 {
	return 1
}

func (d *Disk) createMesh() *polymesh.PolyMesh {

	nv := d.Segments

	dv := 2 * m.Pi / float32(nv)

	msh := polymesh.PolyMesh{NodeDef: d.NodeDef, NodeName: d.NodeName + ":<mesh>",
		IsVisible: true,
		Shader:    d.Shader}

	msh.ModelToWorld.Elems = []m.Matrix4{m.Matrix4Identity}
	msh.ModelToWorld.MotionKeys = 1

	ang := float32(0)

	msh.Verts.MotionKeys = 1
	msh.Normals.MotionKeys = 1

	for i := 0; i < nv; i++ {
		msh.Verts.Elems = append(msh.Verts.Elems, d.P)
		msh.Verts.Elems = append(msh.Verts.Elems, m.Vec3Add(d.P, m.Vec3Add(m.Vec3Scale(d.Radius*m.Cos(ang), d.B), m.Vec3Scale(d.Radius*m.Sin(ang), d.T))))
		msh.Verts.Elems = append(msh.Verts.Elems, m.Vec3Add(d.P, m.Vec3Add(m.Vec3Scale(d.Radius*m.Cos(ang+dv), d.B), m.Vec3Scale(d.Radius*m.Sin(ang+dv), d.T))))
		msh.Verts.ElemsPerKey += 3
		msh.Normals.Elems = append(msh.Normals.Elems, d.N)
		msh.Normals.Elems = append(msh.Normals.Elems, d.N)
		msh.Normals.Elems = append(msh.Normals.Elems, d.N)
		msh.Normals.ElemsPerKey += 3
		msh.UV.Elems = append(msh.UV.Elems, m.Vec2{0.5, 0.5})
		msh.UV.Elems = append(msh.UV.Elems, m.Vec2{0.5 * (1 + m.Cos(ang)), 0.5 * (1 + m.Sin(ang))})
		msh.UV.Elems = append(msh.UV.Elems, m.Vec2{0.5 * (1 + m.Cos(ang+dv)), 0.5 * (1 + m.Sin(ang+dv))})
		msh.UV.ElemsPerKey += 3

		ang += dv
		//face.setup()
		//log.Printf("%v", face)
	}

	return &msh
}

func init() {
	nodes.Register("DiskLight", func() (core.Node, error) {

		return &Disk{Segments: 20}, nil

	})
}
