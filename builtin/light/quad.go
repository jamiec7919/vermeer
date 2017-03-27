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

// Quad represents a (planar) quadrilateral light node.
type Quad struct {
	NodeDef  core.NodeDef `node:"-"`
	NodeName string       `node:"Name"`
	P        m.Vec3
	U, V     m.Vec3
	p        [4]m.Vec3 // Corners
	Shader   string

	Samples int `node:",opt"`

	shader core.Shader
	geom   core.Geom
}

var _ core.Node = (*Quad)(nil)
var _ core.Light = (*Quad)(nil)

// Name implements core.Node.
func (d *Quad) Name() string { return d.NodeName }

// Def implements core.Node.
func (d *Quad) Def() core.NodeDef { return d.NodeDef }

// PreRender implelments core.Node.
func (d *Quad) PreRender() error {

	if s := core.FindNode(d.Shader); s != nil {
		shader, ok := s.(core.Shader)

		if !ok {
			return fmt.Errorf("Unable to find shader %v", d.Shader)
		}

		d.shader = shader

	} else {
		return fmt.Errorf("Unable to find node (shader %v)", d.Shader)

	}

	d.p[0] = d.P
	d.p[1] = m.Vec3Add(d.P, d.U)
	d.p[2] = m.Vec3Add3(d.P, d.U, d.V)
	d.p[3] = m.Vec3Add(d.P, d.V)

	geom := d.createMesh()
	core.AddNode(geom)
	d.geom = geom

	return nil
}

// PostRender implelments core.Node.
func (d *Quad) PostRender() error { return nil }

// PotentialContrib implements core.Light.
func (d *Quad) PotentialContrib(sg *core.ShaderContext) float32 {

	return 0.5
}

// NumSamples implements core.Light
func (d *Quad) NumSamples(sg *core.ShaderContext) int {
	return 1 << uint(d.Samples)
}

// Geom implements core.Light
func (d *Quad) Geom() core.Geom { return d.geom }

// ValidSample implements core.Light.
func (d *Quad) ValidSample(sg *core.ShaderContext, sample *core.BSDFSample) bool {
	panic("light/Quad.ValidSample: unimplemented.")
	return false
}

// SampleArea implements core.Light.
func (d *Quad) SampleArea(sg *core.ShaderContext, n int) error {
	panic("light/Quad.SampleArea: unimplemented.")

	a1 := sphericalTriangleArea(d.p[0], d.p[1], d.p[2], sg.P)
	a2 := sphericalTriangleArea(d.p[0], d.p[2], d.p[3], sg.P)
	w1 := float64(a1 / (a1 + a2))
	w2 := float64(a2 / (a1 + a2))

	for i := 0; i < n; i++ {
		idx := uint64(sg.I*d.NumSamples(sg) + i)
		r0 := ldseq.VanDerCorput(idx, sg.Scramble[0])
		r1 := ldseq.Sobol(idx, sg.Scramble[1])

		var x m.Vec3
		var pdf float64

		if r0 < w1 {
			r0 = r0 / (w1 + w2)
			x, pdf = sampleSphericalTriangle(d.p[0], d.p[1], d.p[2], sg.P, r0, r1)

			pdf *= w1
		} else {
			r0 = (r0 - w1) / (w1 + w2)
			x, pdf = sampleSphericalTriangle(d.p[0], d.p[2], d.p[3], sg.P, r0, r1)
			pdf *= 1 - w1
		}

		// x is point on sphere, do a ray-plane intersection
		N := m.Vec3Normalize(m.Vec3Cross(d.U, d.V))

		planeD := -m.Vec3Dot(N, d.p[0])

		t := -(m.Vec3Dot(sg.P, N) + planeD) / m.Vec3Dot(x, N)

		P := m.Vec3Mad(sg.P, x, t)

		//fmt.Printf("\n%v %v %v %v\n", x, t, N, P)
		D := m.Vec3Sub(P, sg.P)

		var ls core.LightSample

		ls.Ldist = m.Vec3Length(D)
		ls.Ld = m.Vec3Normalize(D)

		if m.Vec3Dot(ls.Ld, N) > 0 {
			// Warn?  This shouldn't happen
			continue
		}

		ls.Liu.Lambda = sg.Lambda

		lsg := sg.NewShaderContext()

		lsg.Lambda = sg.Lambda
		lsg.U = 0
		lsg.V = 0
		lsg.N = N
		lsg.Ng = N
		lsg.P = P

		E := d.shader.EvalEmission(lsg, m.Vec3Neg(ls.Ld))

		sg.ReleaseShaderContext(lsg)
		//		sg.Liu.FromRGB(E[0]*ODotN, E[1]*ODotN, E[2]*ODotN)
		//E.Scale(m.Abs(omegaO[2]))
		ls.Liu.FromRGB(E)

		// geometry term / pdf, lots of cancellations
		// http://www.cs.virginia.edu/~jdl/bib/globillum/mis/shirley96.pdf
		ls.Pdf = float32(pdf) * (ls.Ldist * ls.Ldist) / m.Abs(m.Vec3Dot(ls.Ld, N))

		sg.Lsamples = append(sg.Lsamples, ls)
	}
	return nil
}

// DiffuseShadeMult implements core.Light.
func (d *Quad) DiffuseShadeMult() float32 {
	return 1
}

func (d *Quad) createMesh() *polymesh.PolyMesh {

	msh := polymesh.PolyMesh{NodeDef: d.NodeDef, NodeName: d.NodeName + ":<mesh>",
		IsVisible: true,
		Shader:    []string{d.Shader}}

	//msh.ModelToWorld.Elems = []m.Matrix4{m.Matrix4Identity}
	//msh.ModelToWorld.MotionKeys = 1

	msh.Verts.MotionKeys = 1
	msh.Normals.MotionKeys = 1

	N := m.Vec3Normalize(m.Vec3Cross(d.U, d.V))

	msh.Verts.Elems = append(msh.Verts.Elems, d.p[0])
	msh.Verts.Elems = append(msh.Verts.Elems, d.p[1])
	msh.Verts.Elems = append(msh.Verts.Elems, d.p[2])
	msh.Verts.Elems = append(msh.Verts.Elems, d.p[3])
	msh.Verts.ElemsPerKey += 4
	msh.Normals.Elems = append(msh.Normals.Elems, N)
	msh.Normals.Elems = append(msh.Normals.Elems, N)
	msh.Normals.Elems = append(msh.Normals.Elems, N)
	msh.Normals.Elems = append(msh.Normals.Elems, N)
	msh.Normals.ElemsPerKey += 4
	msh.UV.Elems = append(msh.UV.Elems, m.Vec2{0, 0})
	msh.UV.Elems = append(msh.UV.Elems, m.Vec2{1, 0})
	msh.UV.Elems = append(msh.UV.Elems, m.Vec2{1, 1})
	msh.UV.Elems = append(msh.UV.Elems, m.Vec2{0, 1})
	msh.UV.ElemsPerKey += 4
	msh.PolyCount = []int32{4}
	msh.FaceIdx = []int32{0, 1, 2, 3}
	msh.UVIdx = []int32{0, 1, 2, 3}

	return &msh
}

func init() {
	nodes.Register("QuadLight", func() (core.Node, error) {

		return &Quad{Samples: 1}, nil

	})
}
