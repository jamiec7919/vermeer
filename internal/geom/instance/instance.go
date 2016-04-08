// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package instance

import (
	"errors"
	"github.com/jamiec7919/vermeer/internal/core"
	m "github.com/jamiec7919/vermeer/math"
)

type Instance struct {
	NodeName  string
	Primitive string
	Transform m.Matrix4

	prim         core.Primitive
	invTransform m.Matrix4
}

func (i *Instance) Name() string { return i.NodeName }
func (i *Instance) PreRender(rc *core.RenderContext) error {
	node := rc.FindNode(i.Primitive)

	if prim, ok := node.(core.Primitive); ok {
		i.prim = prim
	} else {
		return core.ErrNotFound
	}

	if invTransform, ok := m.Matrix4Inverse(i.Transform); !ok {
		return errors.New("Instance: matrix singular.")
	} else {

		i.invTransform = invTransform
	}

	return nil
}
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
		ray.Result.Ng = m.Matrix4MulVec(i.Transform, ray.Result.Ng)
		ray.Result.B = m.Matrix4MulVec(i.Transform, ray.Result.B)
		ray.Result.T = m.Matrix4MulVec(i.Transform, ray.Result.T)
		ray.Result.Ns = m.Matrix4MulVec(i.Transform, ray.Result.Ns)
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

func create(rc *core.RenderContext, params core.Params) (interface{}, error) {
	i := Instance{}

	if err := params.Unmarshal(&i); err != nil {
		return nil, err
	}

	return &i, nil
}

func init() {
	core.RegisterType("Instance", create)
}
