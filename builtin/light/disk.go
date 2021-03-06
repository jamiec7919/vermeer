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

// NOTE: Conversion 1/|A| sampling (point on surface) to solid angle pdf is
// p := 1/|A| * (|x-y|^2  / cos omega')

// Disk represents a circular disk light node.
type Disk struct {
	NodeDef       core.NodeDef `node:"-"`
	NodeName      string       `node:"Name"`
	P, Up, LookAt m.Vec3
	T, B, N       m.Vec3 `node:"-"`
	Radius        float32
	Shader        string
	Segments      int `node:",opt"`
	Samples       int `node:",opt"`

	shader core.Shader
	geom   core.Geom
}

var _ core.Node = (*Disk)(nil)
var _ core.Light = (*Disk)(nil)

// https://www.scratchapixel.com/lessons/3d-basic-rendering/minimal-ray-tracer-rendering-simple-shapes/ray-plane-and-ray-disk-intersection
func rayPlaneIntersect(Ro, Rd, P, N m.Vec3) (float32, bool) {
	// assuming vectors are all normalized
	denom := m.Vec3Dot(N, Rd)

	if m.Abs(denom) > 1e-6 {
		p0l0 := m.Vec3Sub(P, Ro)

		t := m.Vec3Dot(p0l0, N) / denom

		return t, t >= 0
	}

	return 0, false
}

// https://www.scratchapixel.com/lessons/3d-basic-rendering/minimal-ray-tracer-rendering-simple-shapes/ray-plane-and-ray-disk-intersection
func rayDiskIntersect(Ro, Rd, P, N m.Vec3, radius float32) (float32, bool) {

	if t, ok := rayPlaneIntersect(Ro, Rd, P, N); ok {
		p := m.Vec3Mad(Ro, Rd, t)

		v := m.Vec3Sub(p, P)

		d2 := m.Vec3Dot(v, v)

		return t, m.Sqrt(d2) <= radius
		// or you can use the following optimisation (and precompute radius^2)
		// return d2 <= radius2; // where radius2 = radius * radius
	}

	return 0, false
}

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
	return 1 << uint(d.Samples)
}

// Geom implements core.Light
func (d *Disk) Geom() core.Geom { return d.geom }

// PostRender implelments core.Node.
func (d *Disk) PostRender() error { return nil }

// ValidSample implements core.Light.
func (d *Disk) ValidSample(sg *core.ShaderContext, sample *core.BSDFSample) bool {

	pdf := float64(1.0 / (m.Pi * d.Radius * d.Radius))

	t, ok := rayDiskIntersect(sg.P, sample.D, d.P, d.N, d.Radius)

	if !ok {
		return false
	}

	p := m.Vec3Mad(sg.P, sample.D, t)

	V := m.Vec3Sub(p, sg.P)

	if m.Vec3Dot(V, sg.Ng) <= 0.0 || m.Vec3Dot(V, d.N) >= 0.0 {
		return false
	}

	sample.Ldist = m.Vec3Length(V)
	sample.Ld = m.Vec3Normalize(V)

	sample.Liu.Lambda = sg.Lambda
	//ODotN := omegaO[2]

	lsg := sg.NewShaderContext()

	lsg.Lambda = sg.Lambda
	lsg.P = p
	lsg.N = d.N
	lsg.Ng = d.N
	lsg.U = m.Vec3Dot(d.B, m.Vec3Sub(p, d.P))
	lsg.V = m.Vec3Dot(d.T, m.Vec3Sub(p, d.P))

	E := d.shader.EvalEmission(lsg, m.Vec3Neg(sample.Ld))

	sg.ReleaseShaderContext(lsg)

	sample.Liu.FromRGB(E)

	// Convert dA to dSigma
	sample.PdfLight = float32(pdf) * (sample.Ldist * sample.Ldist) / (m.Abs(m.Vec3Dot(sample.Ld, d.N)))

	return true
}

// SampleArea implements core.Light.
func (d *Disk) SampleArea(sg *core.ShaderContext, n int) error {
	for i := 0; i < n; i++ {
		idx := uint64(sg.I*n + i)
		r0 := ldseq.VanDerCorput(idx, sg.Scramble[0])
		r1 := ldseq.Sobol(idx, sg.Scramble[1])

		u := d.Radius * m.Sqrt(float32(r0)) * m.Cos(2*m.Pi*float32(r1))
		v := d.Radius * m.Sqrt(float32(r0)) * m.Sin(2*m.Pi*float32(r1))

		pdf := float64(1.0 / (m.Pi * d.Radius * d.Radius))

		P := m.Vec3Add3(d.P, m.Vec3Scale(u, d.B), m.Vec3Scale(v, d.T))

		V := m.Vec3Sub(P, sg.P)

		var ls core.LightSample

		if m.Vec3Dot(V, sg.Ng) > 0.0 && m.Vec3Dot(V, d.N) < 0.0 {
			ls.Ldist = m.Vec3Length(V)
			ls.Ld = m.Vec3Normalize(V)

			ls.Liu.Lambda = sg.Lambda

			//ODotN := omegaO[2]
			lsg := sg.NewShaderContext()

			lsg.Lambda = sg.Lambda
			lsg.P = P
			lsg.N = d.N
			lsg.Ng = d.N
			lsg.U = u
			lsg.V = v

			E := d.shader.EvalEmission(lsg, m.Vec3Neg(ls.Ld))
			//		sg.Liu.FromRGB(E[0]*ODotN, E[1]*ODotN, E[2]*ODotN)
			//E.Scale(m.Abs(omegaO[2]))
			ls.Liu.FromRGB(E)

			sg.ReleaseShaderContext(lsg)

			// geometry term / pdf
			ls.Pdf = float32(pdf) * (ls.Ldist * ls.Ldist) / m.Abs(m.Vec3Dot(ls.Ld, d.N))

			//		fmt.Printf("%v %v %v %v\n", sg.Poffset, sg.P, pdf, sg.Weight)
			sg.Lsamples = append(sg.Lsamples, ls)
		}
	}
	return nil
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
		Shader:    []string{d.Shader}}

	//msh.ModelToWorld.Elems = []m.Matrix4{m.Matrix4Identity}
	//msh.ModelToWorld.MotionKeys = 1

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

		return &Disk{Segments: 20, Samples: 1}, nil

	})
}
