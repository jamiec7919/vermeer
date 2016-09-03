// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package light

import (
	"fmt"
	"github.com/jamiec7919/vermeer/builtin/geom/sphere"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/math/ldseq"
	"github.com/jamiec7919/vermeer/nodes"
)

// Sphere represents a spherical light node.
type Sphere struct {
	NodeDef  core.NodeDef
	NodeName string `node:"Name"`
	P        m.Vec3
	Radius   float32
	Shader   string

	Samples int

	shader core.Shader
	geom   core.Geom
}

var _ core.Node = (*Sphere)(nil)
var _ core.Light = (*Sphere)(nil)

// Name implements core.Node.
func (d *Sphere) Name() string { return d.NodeName }

// Def implements core.Node.
func (d *Sphere) Def() core.NodeDef { return d.NodeDef }

// PreRender implelments core.Node.
func (d *Sphere) PreRender() error {

	if s := core.FindNode(d.Shader); s != nil {
		shader, ok := s.(core.Shader)

		if !ok {
			return fmt.Errorf("Unable to find shader %v", d.Shader)
		}

		d.shader = shader

		geom := &sphere.Sphere{P: d.P, Radius: d.Radius, Shader: d.Shader}
		core.AddNode(geom)
		d.geom = geom

	} else {
		return fmt.Errorf("Unable to find node (shader %v)", d.Shader)

	}

	return nil
}

// PostRender implements core.Node.
func (d *Sphere) PostRender() error { return nil }

// Geom implements core.Light
func (d *Sphere) Geom() core.Geom { return d.geom }

func solveQuadratic(a, b, c float32, x0, x1 *float32) bool {
	discr := b*b - 4*a*c

	if discr < 0 {
		return false
	} else if discr == 0 {
		*x1 = -0.5 * b / a
		*x0 = *x1
	} else {
		var q float32

		if b > 0 {
			q = -0.5 * (b + m.Sqrt(discr))
		} else {
			q = -0.5 * (b - m.Sqrt(discr))
		}
		*x0 = q / a
		*x1 = c / q
	}

	if *x0 > *x1 {
		*x0, *x1 = *x1, *x0
	}

	return true
}

func raySphereIntersect(Ro, Rd, P m.Vec3, radius float32) (float32, bool) {
	// analytic solution
	L := m.Vec3Sub(Ro, P)

	a := m.Vec3Dot(Rd, Rd)

	b := 2 * m.Vec3Dot(Rd, L)

	c := m.Vec3Dot(L, L) - radius*radius

	var t0, t1 float32

	if !solveQuadratic(a, b, c, &t0, &t1) {
		return 0, false
	}

	if t0 > t1 {
		t0, t1 = t1, t0
	}

	if t0 < 0 {
		t0 = t1 // if t0 is negative, let's use t1 instead
		if t0 < 0 {
			return 0, false
		} // both t0 and t1 are negative
	}

	return t0, true
}

func sqr(x float32) float32 { return x * x }

// SampleArea implements core.Light.
func (d *Sphere) SampleArea(sg *core.ShaderContext, n int) error {

	V := m.Vec3Sub(d.P, sg.P)

	l := m.Vec3Length(V)

	w := m.Vec3Normalize(V)
	v := m.Vec3Normalize(m.Vec3Cross(w, sg.Ng))
	u := m.Vec3Cross(w, v)

	for i := 0; i < n; i++ {
		idx := uint64(sg.I*sg.NSamples + sg.Sample + i)
		r0 := ldseq.VanDerCorput(idx, sg.SampleScramble)
		r1 := ldseq.Sobol(idx, sg.SampleScramble2)

		theta := m.Acos(1 - float32(r0) + float32(r0)*m.Sqrt(1-sqr(d.Radius/l)))
		phi := 2 * m.Pi * float32(r1)

		a := m.Vec3{
			m.Cos(phi) * m.Sin(theta),
			m.Sin(phi) * m.Sin(theta),
			m.Cos(theta),
		}

		omega := m.Vec3BasisExpand(u, v, w, a)

		if m.Vec3Dot(omega, sg.Ng) < 0 {
			//return fmt.Errorf("nosample")
			continue
		}

		// Intersect ray x+ta with sphere.
		t, ok := raySphereIntersect(sg.P, omega, d.P, d.Radius)

		if !ok {
			//return fmt.Errorf("nosample")
			continue
		}

		x := m.Vec3Mad(sg.P, omega, t)

		D := m.Vec3Sub(x, sg.P)

		var ls core.LightSample

		ls.Ldist = m.Vec3Length(D)
		ls.Ld = m.Vec3Normalize(D)

		ls.Liu.Lambda = sg.Lambda

		lsg := sg.NewShaderContext()
		lsg.Lambda = sg.Lambda

		N := m.Vec3Normalize(m.Vec3Sub(x, d.P))

		lsg.U = 0.5 + m.Atan2(N[2], N[0])/(2*m.Pi)
		lsg.V = 0.5 - m.Asin(N[1])/m.Pi
		lsg.Shader = d.shader

		E := d.shader.EvalEmission(lsg, m.Vec3Neg(omega))

		sg.ReleaseShaderContext(lsg)

		//		sg.Liu.FromRGB(E[0]*ODotN, E[1]*ODotN, E[2]*ODotN)
		//E.Scale(m.Abs(omegaO[2]))
		ls.Liu.FromRGB(E)

		// geometry term / pdf, lots of cancellations
		// http://www.cs.virginia.edu/~jdl/bib/globillum/mis/shirley96.pdf
		ls.Weight = m.Abs(m.Vec3Dot(ls.Ld, sg.N)) * 2 * m.Pi * (1 - m.Sqrt(1-sqr(d.Radius/l)))

		sg.Lsamples = append(sg.Lsamples, ls)
	}
	return nil
}

// NumSamples implements core.Light
func (d *Sphere) NumSamples(sg *core.ShaderContext) int {
	return d.Samples * d.Samples
}

// PotentialContrib implements core.Light.
func (d *Sphere) PotentialContrib(sg *core.ShaderContext) float32 {

	// Should check for plane horizon

	D := m.Vec3Sub(d.P, sg.P)

	ld := m.Vec3Length(D)

	thetaMax := m.Asin(d.Radius / ld)

	//l := m.Vec3Length(D) - d.Radius // Distance to closest point.

	//return m.Sqrt(1 - sqr(d.Radius/l))
	return m.Pi * sqr(m.Sin(thetaMax))
}

// DiffuseShadeMult implements core.Light.
func (d *Sphere) DiffuseShadeMult() float32 {
	return 1
}

func init() {
	nodes.Register("SphereLight", func() (core.Node, error) {

		return &Sphere{Radius: 1, Samples: 1}, nil

	})
}
