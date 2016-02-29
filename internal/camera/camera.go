// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package camera

import (
	"github.com/jamiec7919/vermeer/internal/core"
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

type Camera struct {
	Eye           m.Vec3
	U, V, W       m.Vec3
	L, R, T, B, D float32
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
	c.D = 2.5
	//c.D = 3.5
	// log.Printf("Lookat: %v %v %v", c.U, c.V, c.W)
}

func (c *Camera) ComputeRay(u, v float32, rnd *rand.Rand) (P, D m.Vec3) {
	P = c.Eye
	// D = || u*U + v*V - d*W  ||

	s := m.Vec3Sub(m.Vec3Add(m.Vec3Scale(u, c.U), m.Vec3Scale(v, c.V)), m.Vec3Scale(c.D, c.W))
	D = m.Vec3Normalize(s)

	D = D

	//	log.Printf("%v %v %v %v", D, u, v, vm.Vec3Add(vm.Vec3Scale(u, c.U), vm.Vec3Scale(v, c.V)))
	return
}

type Lookat struct {
	From, To, Up m.Vec3
}

func init() {
	core.RegisterNodeType("Camera", func(rc *core.RenderContext, params core.Params) error {
		var l Lookat
		if err := params.Unmarshal(&l); err != nil {
			return err
		}

		cam := &Camera{}
		cam.Lookat(l.From, l.To, l.Up)
		rc.AddNode(cam)
		return nil
	})
}
