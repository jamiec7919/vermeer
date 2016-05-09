// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package instance

import (
	"errors"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/nodes"
)

type Instance struct {
	NodeName  string
	Primitive string
	Transform m.Matrix4

	prim              core.Primitive
	invTransform      m.Matrix4
	invTransTransform m.Matrix4 // INverse transpose for normals
}

func (i *Instance) Name() string { return i.NodeName }
func (i *Instance) PreRender(rc *core.RenderContext) error {
	node := rc.FindNode(i.Primitive)

	if prim, ok := node.(core.Primitive); ok {
		i.prim = prim
	} else {
		return errors.New("Can't find primitive " + i.Primitive)
	}

	if invTransform, ok := m.Matrix4Inverse(i.Transform); !ok {
		return errors.New("Instance: matrix singular.")
	} else {

		i.invTransform = invTransform
		invTransTransform, _ := m.Matrix4Inverse(m.Matrix4Transpose(i.Transform))
		i.invTransTransform = invTransTransform
	}

	return nil
}

func (i *Instance) Visible() bool                           { return true }
func (i *Instance) PostRender(rc *core.RenderContext) error { return nil }

func (i *Instance) WorldBounds() (out m.BoundingBox) {
	out.Reset()
	box := i.prim.WorldBounds()

	{
		v := m.Matrix4MulPoint(i.Transform, m.Vec3{box.Bounds[0][0], box.Bounds[0][1], box.Bounds[0][2]})
		out.GrowVec3(v)
	}
	{
		v := m.Matrix4MulPoint(i.Transform, m.Vec3{box.Bounds[1][0], box.Bounds[0][1], box.Bounds[0][2]})
		out.GrowVec3(v)
	}
	{
		v := m.Matrix4MulPoint(i.Transform, m.Vec3{box.Bounds[0][0], box.Bounds[1][1], box.Bounds[0][2]})
		out.GrowVec3(v)
	}
	{
		v := m.Matrix4MulPoint(i.Transform, m.Vec3{box.Bounds[0][0], box.Bounds[0][1], box.Bounds[1][2]})
		out.GrowVec3(v)
	}
	{
		v := m.Matrix4MulPoint(i.Transform, m.Vec3{box.Bounds[1][0], box.Bounds[1][1], box.Bounds[0][2]})
		out.GrowVec3(v)
	}
	{
		v := m.Matrix4MulPoint(i.Transform, m.Vec3{box.Bounds[1][0], box.Bounds[0][1], box.Bounds[1][2]})
		out.GrowVec3(v)
	}
	{
		v := m.Matrix4MulPoint(i.Transform, m.Vec3{box.Bounds[0][0], box.Bounds[1][1], box.Bounds[1][2]})
		out.GrowVec3(v)
	}
	{
		v := m.Matrix4MulPoint(i.Transform, m.Vec3{box.Bounds[1][0], box.Bounds[1][1], box.Bounds[1][2]})
		out.GrowVec3(v)
	}

	return
}

func (i *Instance) TraceRay(ray *core.RayData) {
	ray.SavedRay = ray.Ray

	ray.InitRay(m.Matrix4MulPoint(i.invTransform, ray.Ray.P), m.Matrix4MulVec(i.invTransform, ray.Ray.D))

	i.prim.TraceRay(ray)

	if ray.Ray.Tclosest < ray.SavedRay.Tclosest {
		t := ray.Ray.Tclosest
		ray.Ray = ray.SavedRay
		ray.Ray.Tclosest = t

		ray.Result.P = m.Matrix4MulPoint(i.Transform, ray.Result.P)
		ray.Result.POffset = m.Matrix4MulVec(i.Transform, ray.Result.POffset)
		for k := range ray.Result.UV {
			ray.Result.Pu[k] = m.Matrix4MulVec(i.invTransTransform, ray.Result.Pu[k])
			ray.Result.Pv[k] = m.Matrix4MulVec(i.invTransTransform, ray.Result.Pv[k])
		}
		ray.Result.Ng = m.Matrix4MulVec(i.invTransTransform, ray.Result.Ng)
		ray.Result.B = m.Matrix4MulVec(i.invTransTransform, ray.Result.B)
		ray.Result.T = m.Matrix4MulVec(i.invTransTransform, ray.Result.T)
		ray.Result.Ns = m.Matrix4MulVec(i.invTransTransform, ray.Result.Ns)
	} else {
		ray.Ray = ray.SavedRay
	}
}

func (i *Instance) VisRay(ray *core.RayData) {
	ray.SavedRay = ray.Ray

	ray.InitRay(m.Matrix4MulPoint(i.invTransform, ray.Ray.P), m.Matrix4MulVec(i.invTransform, ray.Ray.D))

	i.prim.VisRay(ray)

	if ray.Ray.Tclosest < ray.SavedRay.Tclosest {
		t := ray.Ray.Tclosest
		ray.Ray = ray.SavedRay
		ray.Ray.Tclosest = t
	} else {
		ray.Ray = ray.SavedRay
	}

}

func create() (core.Node, error) {
	i := Instance{}

	return &i, nil
}

func init() {
	nodes.Register("Instance", create)
}
