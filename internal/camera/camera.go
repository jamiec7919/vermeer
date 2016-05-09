// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package camera

import (
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/math/sample"
	"github.com/jamiec7919/vermeer/nodes"
	"math/rand"
)

/*
Need to be able to sample the lens, with a pinhole we have:
Spatial
We(y0) = 1
P_A(y0) = 1

Directional
We(y0->y1) = d^2 / (A_p * cos^3	theta_o),
P(y0->y1) = d^2 / (A_p * cos^3	theta_o),

cos theta_o = D dot camera.N  (camera.N = -camera.W)
A_p is pixel density of image with respect to screen size.

For a circular thin lense model camera:
Spatial
We(y0) = 1 / (pi * r^2)
P_A(y0) = 1 / (pi * r^2)

r is radius of lens/apeture
Angular
(same as pinhole)

Should sample position on lens, and directional component.
Note that for bpt these should be seperate as might want to join to
lens.  However this should mean that light paths can contribute to
*any* pixel when joining to lens. (and indeed may be splatted if
there is a multi-pixel filter)
*/
type Camera struct {
	Eye           m.Vec3
	U, V, W       m.Vec3
	L, R, T, B, D float32
	Radius        float32
}

func (c *Camera) PreRender(*core.RenderContext) error  { return nil }
func (c *Camera) PostRender(*core.RenderContext) error { return nil }
func (c *Camera) Name() string                         { return "camera" }

func (c *Camera) Lookat(e, t, u m.Vec3) {
	c.Eye = e

	c.W = m.Vec3Normalize(m.Vec3Sub(e, t)) // Note W points away from target
	c.U = m.Vec3Normalize(m.Vec3Cross(u, c.W))
	c.V = m.Vec3Normalize(m.Vec3Cross(c.U, c.W))
	//c.D = 1.5
	//c.D = 2.5
	//c.D = 3.5
	// log.Printf("Lookat: %v %v %v", c.U, c.V, c.W)
}

/*
	s := vm.Vec3Add(c.Eye, vm.Vec3Sub(vm.Vec3Add(vm.Vec3Scale(u, c.U), vm.Vec3Scale(v, c.V)), vm.Vec3Scale(c.D, c.W)))

	x, y := sample.UniformDisk2D(0.01, r0, r1)
	e := vm.Vec3Add(c.Eye, vm.Vec3Add(vm.Vec3Scale(x, c.U), vm.Vec3Scale(y, c.V)))
	D := vm.Vec3Normalize(vm.Vec3Sub(s, e))

*/

func (c *Camera) SampleLensArea(rnd *rand.Rand, P *m.Vec3, We *float32, pdf *float32) error {
	*P = c.Eye

	if c.Radius > 0.0 {
		x, y := sample.UniformDisk2D(c.Radius, rnd.Float32(), rnd.Float32())
		e := m.Vec3Add(m.Vec3Scale(x, c.U), m.Vec3Scale(y, c.V))
		*P = m.Vec3Add(*P, e)

		*We = 1.0 / (m.Pi * c.Radius * c.Radius)
		*pdf = 1.0 / (m.Pi * c.Radius * c.Radius)
	} else {
		*We = 1
		*pdf = 1
	}

	return nil
}

func (c *Camera) SampleImagePlaneDir(u, v float32, P m.Vec3, rnd *rand.Rand, omega_o *m.Vec3, We *float32, pdf *float32) error {
	s := m.Vec3Sub(m.Vec3Add(m.Vec3Scale(u, c.U), m.Vec3Scale(v, c.V)), m.Vec3Scale(c.D, c.W))
	*omega_o = m.Vec3Normalize(m.Vec3Sub(m.Vec3Add(c.Eye, s), P))

	cos_omega_o := m.Vec3Dot(*omega_o, m.Vec3Neg(c.W))
	A_p := float32(1.0) // pixel density WRT image size
	*We = (c.D * c.D) / (A_p * cos_omega_o * cos_omega_o * cos_omega_o)
	*pdf = (c.D * c.D) / (A_p * cos_omega_o * cos_omega_o * cos_omega_o)
	return nil
}

func (c *Camera) ComputeRay(u, v float32, rnd *rand.Rand) (P, D m.Vec3) {
	P = c.Eye
	// D = || u*U + v*V - d*W  ||

	s := m.Vec3Sub(m.Vec3Add(m.Vec3Scale(u, c.U), m.Vec3Scale(v, c.V)), m.Vec3Scale(c.D, c.W))

	if c.Radius > 0.0 {
		x, y := sample.UniformDisk2D(c.Radius, rnd.Float32(), rnd.Float32())
		e := m.Vec3Add(m.Vec3Scale(x, c.U), m.Vec3Scale(y, c.V))
		D = m.Vec3Normalize(m.Vec3Sub(s, e))
		P = m.Vec3Add(P, e)
	} else {
		D = m.Vec3Normalize(s)
	}
	//D = D

	//	log.Printf("%v %v %v %v", D, u, v, vm.Vec3Add(vm.Vec3Scale(u, c.U), vm.Vec3Scale(v, c.V)))
	return
}

type Lookat struct {
	From, To, Up m.Vec3
	D, Radius    float32
}

func init() {
	nodes.Register("Camera", func() (core.Node, error) {
		l := Lookat{D: 2.5}

		cam := &Camera{D: l.D, Radius: l.Radius}
		cam.Lookat(l.From, l.To, l.Up)
		//rc.AddNode(cam)
		return cam, nil
	})
}
