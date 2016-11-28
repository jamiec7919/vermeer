// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package instance

import (
	"fmt"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/core/param"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/nodes"
)

type TransformSRTArray []m.TransformDecomp

func (t *TransformSRTArray) TimeKey(time float32) m.TransformDecomp {
	if len(*t) > 1 {
		k := time * float32(len(*t)-1)

		timeFrac := k - m.Floor(k)

		key := int(m.Floor(k))
		key2 := int(m.Ceil(k))

		if key > len(*t)-1 || key2 > len(*t)-1 {
			panic(fmt.Sprintf("TransformSRTArray.TimeKey: %v %v %v", time, key, key2))
		}
		//fmt.Printf("%v %v %v %v %v %v %v", ray.Time, len(mesh.Transform.Elems), time, key, key2, len(mesh.transformSRT), mesh.transformSRT)

		return m.TransformDecompLerp((*t)[key], (*t)[key2], timeFrac)
	}

	return (*t)[0]
}

// Instance duplicates an existing geom but with a new transform.
type Instance struct {
	NodeDef  core.NodeDef `node:"-"`
	NodeName string       `node:"Name"`

	Geom string

	BMin param.PointArray
	BMax param.PointArray

	Transform    param.MatrixArray
	transformSRT TransformSRTArray

	geom core.Geom

	bounds []m.BoundingBox
}

// Assert that Instance implements important interfaces.
var _ core.Node = (*Instance)(nil)
var _ core.Geom = (*Instance)(nil)

// Name is a core.Node method.
func (ins *Instance) Name() string { return ins.NodeName }

// Def is a core.Node method.
func (ins *Instance) Def() core.NodeDef { return ins.NodeDef }

// Bounds implements core.Geom.
func (ins *Instance) Bounds(time float32) m.BoundingBox {
	// This is reasonable but should take from the actual given bounds

	return ins.bounds[0]
}

// Trace implements core.Primitive.
func (ins *Instance) Trace(ray *core.Ray, sg *core.ShaderContext) bool {
	// At this point should transform ray.
	var Rp, Rd, Rdinv m.Vec3
	var S [3]float32
	var Kx, Ky, Kz int32
	var transform, invTransform m.Matrix4

	Rp = ray.P
	Rd = ray.D
	Rdinv = ray.Dinv
	S = ray.S
	Kx = ray.Kx
	Ky = ray.Ky
	Kz = ray.Kz

	//			transformSRT := m.TransformDecompLerp(mesh.transformSRT[key], mesh.transformSRT[key2], time)
	transform = m.TransformDecompToMatrix4(ins.transformSRT.TimeKey(ray.Time))

	invTransform, _ = m.Matrix4Inverse(transform)

	ray.P = m.Matrix4MulPoint(invTransform, Rp)
	ray.D = m.Matrix4MulVec(invTransform, Rd)
	ray.Setup()

	hit := ins.geom.Trace(ray, sg)

	ray.P = Rp
	ray.D = Rd
	ray.Dinv = Rdinv
	ray.S = S
	ray.Kx = Kx
	ray.Ky = Ky
	ray.Kz = Kz

	if hit {
		// At this point we know that this ray has hit this geom as closest point so ok to overwrite.
		sg.Transform = transform
		sg.InvTransform = invTransform
	}

	return hit
}

// PreRender is a core.Node method.
func (ins *Instance) PreRender() error {

	for i := range ins.Transform.Elems {
		ins.transformSRT = append(ins.transformSRT, m.TransformDecompMatrix4(ins.Transform.Elems[i]))
	}

	for i := range ins.BMin.Elems {
		box := m.BoundingBox{}

		box.GrowVec3(ins.BMin.Elems[i])
		box.GrowVec3(ins.BMax.Elems[i])

		ins.bounds = append(ins.bounds, box)
	}

	if s := core.FindNode(ins.Geom); s != nil {
		geom, ok := s.(core.Geom)

		if !ok {
			return fmt.Errorf("Instance %v: Unable to find geom %v", ins.Geom)
		}

		ins.geom = geom
	} else {
		return fmt.Errorf("Instance %v: Unable to find node %v", ins.Geom)

	}

	return nil
}

// PostRender is a core.Node method.
func (ins *Instance) PostRender() error { return nil }

// MotionKeys returns the number of motion keys.
func (ins *Instance) MotionKeys() int {

	return ins.geom.MotionKeys()

}

func create() (core.Node, error) {
	mfile := Instance{}

	return &mfile, nil
}

func init() {
	nodes.Register("GeomInstance", create)
}
