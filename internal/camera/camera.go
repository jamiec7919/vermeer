// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package camera

import (
	"github.com/jamiec7919/vermeer/internal/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/math/sample"
	"math/rand"
)

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
	core.RegisterNodeType("Camera", func(rc *core.RenderContext, params core.Params) error {
		l := Lookat{D: 2.5}
		if err := params.Unmarshal(&l); err != nil {
			return err
		}

		cam := &Camera{D: l.D, Radius: l.Radius}
		cam.Lookat(l.From, l.To, l.Up)
		rc.AddNode(cam)
		return nil
	})
}
