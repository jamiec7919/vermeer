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

// Instance represents a transformed instance of an existing primitive node.
type Instance struct {
	NodeName  string
	Primitive string
	Transform m.Matrix4

	prim              core.Primitive
	invTransform      m.Matrix4
	invTransTransform m.Matrix4 // INverse transpose for normals
}

// Assert that Instance implements the important interfaces.
var _ core.Node = (*Instance)(nil)
var _ core.Primitive = (*Instance)(nil)

// Name implements core.Node.
func (i *Instance) Name() string { return i.NodeName }

// PreRender implements core.Node.
func (i *Instance) PreRender(rc *core.RenderContext) error {
	node := rc.FindNode(i.Primitive)

	if prim, ok := node.(core.Primitive); ok {
		i.prim = prim
	} else {
		return errors.New("Can't find primitive " + i.Primitive)
	}

	if invTransform, ok := m.Matrix4Inverse(i.Transform); ok {
		i.invTransform = invTransform
		invTransTransform, _ := m.Matrix4Inverse(m.Matrix4Transpose(i.Transform))
		i.invTransTransform = invTransTransform
	} else {
		return errors.New("Instance: matrix singular.")

	}

	return nil
}

// Visible implements core.Primitive.
func (i *Instance) Visible() bool { return true }

// PostRender implements core.Node.
func (i *Instance) PostRender(rc *core.RenderContext) error { return nil }

// WorldBounds implements core.Primtive.
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

// TraceRay implements core.Primitive.
func (i *Instance) TraceRay(ray *core.RayData, sg *core.ShaderGlobals) (mtlid int32) {
	ray.SavedRay = ray.Ray

	//	ray.InitRay(m.Matrix4MulPoint(i.invTransform, ray.Ray.P), m.Matrix4MulVec(i.invTransform, ray.Ray.D))

	mtlid = i.prim.TraceRay(ray, sg)

	if ray.Ray.Tclosest < ray.SavedRay.Tclosest {
		t := ray.Ray.Tclosest
		ray.Ray = ray.SavedRay
		ray.Ray.Tclosest = t

		ray.Result.P = m.Matrix4MulPoint(i.Transform, ray.Result.P)
		ray.Result.POffset = m.Matrix4MulVec(i.Transform, ray.Result.POffset)
		ray.Result.Pu = m.Matrix4MulVec(i.invTransTransform, ray.Result.Pu)
		ray.Result.Pv = m.Matrix4MulVec(i.invTransTransform, ray.Result.Pv)
		ray.Result.Ng = m.Matrix4MulVec(i.invTransTransform, ray.Result.Ng)
		ray.Result.B = m.Matrix4MulVec(i.invTransTransform, ray.Result.B)
		ray.Result.T = m.Matrix4MulVec(i.invTransTransform, ray.Result.T)
		ray.Result.Ns = m.Matrix4MulVec(i.invTransTransform, ray.Result.Ns)
	} else {
		ray.Ray = ray.SavedRay
	}

	return
}

// VisRay implements core.Primitive.
func (i *Instance) VisRay(ray *core.RayData) {
	ray.SavedRay = ray.Ray

	ray.Init(core.RayShadow, m.Matrix4MulPoint(i.invTransform, ray.Ray.P), m.Matrix4MulVec(i.invTransform, ray.Ray.D), 1, &core.ShaderGlobals{})

	//i.prim.VisRay(ray)
	core.TraceProbe(ray, &core.ShaderGlobals{})
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
