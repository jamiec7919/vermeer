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

// Tri represents a triangular light node.
type Tri struct {
	NodeDef    core.NodeDef `node:"-"`
	NodeName   string       `node:"Name"`
	P0, P1, P2 m.Vec3
	Shader     string

	shader core.Shader
	geom   core.Geom
}

// Name implements core.Node.
func (d *Tri) Name() string { return d.NodeName }

// Def implements core.Node.
func (d *Tri) Def() core.NodeDef { return d.NodeDef }

// PreRender implelments core.Node.
func (d *Tri) PreRender() error {

	if s := core.FindNode(d.Shader); s != nil {
		shader, ok := s.(core.Shader)

		if !ok {
			return fmt.Errorf("Unable to find shader %v", d.Shader)
		}

		d.shader = shader

		geom := d.createMesh()
		core.AddNode(geom)
		d.geom = geom

	} else {
		return fmt.Errorf("Unable to find node (shader %v)", d.Shader)

	}

	return nil
}

// PostRender implelments core.Node.
func (d *Tri) PostRender() error { return nil }

// PotentialContrib imeplements core.Light
func (d *Tri) PotentialContrib(sg *core.ShaderContext) float32 {
	area := sphericalTriangleArea(d.P0, d.P1, d.P2, sg.P)

	return 1 - m.Log2(area)
}

// NumSamples implements core.Light
func (d *Tri) NumSamples(sg *core.ShaderContext) int {
	return 1
}

// Geom implements core.Light
func (d *Tri) Geom() core.Geom { return d.geom }

// SampleArea implements core.Light.
func (d *Tri) SampleArea(sg *core.ShaderContext, n int) error {

	for i := 0; i < n; i++ {
		idx := uint64(sg.I*d.NumSamples(sg) + sg.Sample + i)
		r0 := ldseq.VanDerCorput(idx, sg.Scramble[0])
		r1 := ldseq.Sobol(idx, sg.Scramble[1])

		x, pdf := sampleSphericalTriangle(d.P0, d.P1, d.P2, sg.P, r0, r1)

		// x is point on sphere, do a ray-plane intersection
		N := m.Vec3Normalize(m.Vec3Cross(m.Vec3Sub(d.P2, d.P0), m.Vec3Sub(d.P1, d.P0)))

		planeD := -m.Vec3Dot(N, d.P0)

		t := -(m.Vec3Dot(sg.P, N) + planeD) / m.Vec3Dot(x, N)

		P := m.Vec3Mad(sg.P, x, t)

		//fmt.Printf("\n%v %v %v %v\n", x, t, N, P)
		D := m.Vec3Sub(P, sg.P)

		var ls core.LightSample

		ls.Ldist = m.Vec3Length(D)
		ls.Ld = m.Vec3Normalize(D)

		if m.Vec3Dot(ls.Ld, N) > 0 {
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
		ls.Weight = m.Abs(m.Vec3Dot(ls.Ld, sg.N)) / /* m.Abs(m.Vec3Dot(sg.Ld, N)) / (sg.Ldist * sg.Ldist */ float32(pdf)

		sg.Lsamples = append(sg.Lsamples, ls)
	}
	return nil
}

// DiffuseShadeMult implements core.Light.
func (d *Tri) DiffuseShadeMult() float32 {
	return 1
}

func init() {
	nodes.Register("TriLight", func() (core.Node, error) {

		return &Tri{}, nil

	})
}

//https://people.sc.fsu.edu/~jburkardt/c_src/sphere_triangle_monte_carlo/sphere_triangle_monte_carlo.c
//http://www.graphics.cornell.edu/pubs/1995/Arv95c.pdf

// Bhaskara I's sine approximation formula
// Valid [0,pi]
func sinapprox(x float32) float32 {
	num := 16 * x * (m.Pi - x)
	denom := 5*m.Pi*m.Pi - 4*x*(m.Pi-x)
	return num / denom
}

// http://nghiaho.com/?p=997
func atanapprox(x float32) float32 {
	return m.Pi*x - x*(m.Abs(x)-1)*(0.2447+0.0663*m.Abs(x))
}

// triangleSidesToAngles takes the geodesic lengths of the sides of a spherical
// triangle and calculates the angles of the triangle.
func triangleSidesToAngles(as, bs, cs float32) (a, b, c float32) {
	ssu := (as + bs + cs) / 2

	/*
		tanA2 := m.Sqrt((m.Sin(ssu-bs) * m.Sin(ssu-cs)) / (m.Sin(ssu) * m.Sin(ssu-as)))

		a = 2 * m.Atan(tanA2)

		tanB2 := m.Sqrt((m.Sin(ssu-as) * m.Sin(ssu-cs)) / (m.Sin(ssu) * m.Sin(ssu-bs)))

		b = 2 * m.Atan(tanB2)

		tanC2 := m.Sqrt((m.Sin(ssu-as) * m.Sin(ssu-bs)) / (m.Sin(ssu) * m.Sin(ssu-cs)))

		c = 2 * m.Atan(tanC2)
	*/

	sinssuas := m.Sin(ssu - as)
	sinssubs := m.Sin(ssu - bs)
	sinssucs := m.Sin(ssu - cs)
	sinssu := m.Sin(ssu)
	/*
		sinssuas := sinapprox(ssu - as)
		sinssubs := sinapprox(ssu - bs)
		sinssucs := sinapprox(ssu - cs)
		sinssu := sinapprox(ssu)
	*/
	tanA2 := m.Sqrt(sinssubs * sinssucs / (sinssu * sinssuas))
	tanB2 := m.Sqrt(sinssuas * sinssucs / (sinssu * sinssubs))
	tanC2 := m.Sqrt(sinssuas * sinssubs / (sinssu * sinssucs))

	//fmt.Printf("%v %v %v\n", tanA2, tanB2, tanC2)

	a = 2 * m.Atan(tanA2)
	b = 2 * m.Atan(tanB2)
	c = 2 * m.Atan(tanC2)
	/*
		a = 2 * atanapprox(tanA2)
		b = 2 * atanapprox(tanB2)
		c = 2 * atanapprox(tanC2)
	*/
	return
}

//go:nosplit
func acosapprox(x float32) float32 {
	a := m.Sqrt(2 + 2*x)
	b := m.Sqrt(2 - 2*x)
	c := m.Sqrt(2 - a)
	return 8/3*c - b/3

}

//go:nosplit
func triangleVerticesToSides(pa, pb, pc m.Vec3) (a, b, c float32) {

	a = m.Acos(m.Vec3Dot(pb, pc))
	b = m.Acos(m.Vec3Dot(pc, pa))
	c = m.Acos(m.Vec3Dot(pa, pb))
	/*
		a = acosapprox(m.Vec3Dot(pb, pc))
		b = acosapprox(m.Vec3Dot(pc, pa))
		c = acosapprox(m.Vec3Dot(pa, pb))
	*/
	return
}

func sphericalTriangleArea(p0, p1, p2, p m.Vec3) float32 {
	pa := m.Vec3Normalize(m.Vec3Sub(p0, p))
	pb := m.Vec3Normalize(m.Vec3Sub(p1, p))
	pc := m.Vec3Normalize(m.Vec3Sub(p2, p))

	a, b, c := triangleVerticesToSides(pa, pb, pc)

	alpha, beta, gamma := triangleSidesToAngles(a, b, c)

	area := alpha + beta + gamma - m.Pi
	return area
}

func sampleSphericalTriangle(p0, p1, p2, p m.Vec3, r0, r1 float64) (m.Vec3, float64) {
	pa := m.Vec3Normalize(m.Vec3Sub(p0, p))
	pb := m.Vec3Normalize(m.Vec3Sub(p1, p))
	pc := m.Vec3Normalize(m.Vec3Sub(p2, p))

	a, b, c := triangleVerticesToSides(pa, pb, pc)

	alpha, beta, gamma := triangleSidesToAngles(a, b, c)

	area := alpha + beta + gamma - m.Pi

	areaHat := float32(r0) * area

	//	s := m.Sin(areaHat - alpha)
	//	t := m.Cos(areaHat - alpha)
	s, t := m.Sincos(areaHat - alpha)

	sinAlpha, cosAlpha := m.Sincos(alpha)
	//u := t - m.Cos(alpha)
	//v := s + m.Sin(alpha)*m.Cos(c)
	u := t - cosAlpha
	v := s + sinAlpha*m.Cos(c)

	//	q := ((v*t-u*s)*m.Cos(alpha) - v) / ((v*s + u*t) * m.Sin(alpha))
	q := ((v*t-u*s)*cosAlpha - v) / ((v*s + u*t) * sinAlpha)

	// Possibly out of bounds so clamp
	q = m.Max(-1, m.Min(q, 1))

	w := m.Vec3Dot(pc, pa)

	var v31 m.Vec3
	for k := range v31 {
		v31[k] = pc[k] - w*pa[k]
	}

	v31 = m.Vec3Normalize(v31)

	var v4 m.Vec3

	for k := range v4 {
		v4[k] = q*pa[k] + m.Sqrt(1-q*q)*v31[k]
	}

	z := 1 - float32(r1)*(1-m.Vec3Dot(v4, pb))

	w = m.Vec3Dot(v4, pb)

	var v42 m.Vec3

	for k := range v42 {
		v42[k] = v4[k] - w*pb[k]
	}

	v42 = m.Vec3Normalize(v42)

	x := m.Vec3Add(m.Vec3Scale(z, pb), m.Vec3Scale(m.Sqrt(1-z*z), v42))

	//fmt.Printf("%v %v %v", x, area, areaHat)
	// x is unit vector (point on sphere) - should project this onto triangle plane
	return x, 1 / float64(area)
}

func (d *Tri) createMesh() *polymesh.PolyMesh {

	msh := polymesh.PolyMesh{NodeDef: d.NodeDef, NodeName: d.NodeName + ":<mesh>",
		IsVisible: true,
		Shader:    []string{d.Shader}}

	//msh.ModelToWorld.Elems = []m.Matrix4{m.Matrix4Identity}
	//msh.ModelToWorld.MotionKeys = 1

	msh.Verts.MotionKeys = 1
	msh.Normals.MotionKeys = 1

	N := m.Vec3Normalize(m.Vec3Cross(m.Vec3Sub(d.P2, d.P0), m.Vec3Sub(d.P1, d.P0)))

	msh.Verts.Elems = append(msh.Verts.Elems, d.P0)
	msh.Verts.Elems = append(msh.Verts.Elems, d.P1)
	msh.Verts.Elems = append(msh.Verts.Elems, d.P2)
	msh.Verts.ElemsPerKey += 3
	msh.Normals.Elems = append(msh.Normals.Elems, N)
	msh.Normals.Elems = append(msh.Normals.Elems, N)
	msh.Normals.Elems = append(msh.Normals.Elems, N)
	msh.Normals.ElemsPerKey += 3
	msh.UV.Elems = append(msh.UV.Elems, m.Vec2{0, 0})
	msh.UV.Elems = append(msh.UV.Elems, m.Vec2{1, 0})
	msh.UV.Elems = append(msh.UV.Elems, m.Vec2{0, 1})
	msh.UV.ElemsPerKey += 3

	return &msh
}
