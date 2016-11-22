// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proc

import (
	"fmt"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/core/param"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/nodes"
)

// TransformSRTArray is an array of Transforms supporting lerp.
type TransformSRTArray []m.TransformDecomp

// TimeKey returns the interpolated transform from the array (whole array is assumed to cover [0,1) linearly.
// This should be altered to allow time to be non-linearly related to the keys.
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

// Handler is the type of procedural generation/loading handlers.
type Handler interface {
	Init(proc *Proc, datastring string, userdata interface{}) error
}

// Proc supports procedural loading/generation of geometry via handlers. Handlers may create Geom or Shader nodes.
type Proc struct {
	NodeDef  core.NodeDef `node:"-"`
	NodeName string       `node:"Name"`

	Handler string
	handler Handler

	DataString string      `node:"Data"`
	Userdata   interface{} `node:"-"`

	Geom   []core.Geom   `node:"-"`
	Shader []core.Shader `node:"-"`

	BMin param.PointArray
	BMax param.PointArray

	Transform    param.MatrixArray `node:",opt"`
	transformSRT TransformSRTArray

	//geom core.Geom

	bounds []m.BoundingBox
}

// Assert that Proc implements important interfaces.
var _ core.Node = (*Proc)(nil)
var _ core.Geom = (*Proc)(nil)

// Name is a core.Node method.
func (proc *Proc) Name() string { return proc.NodeName }

// Def is a core.Node method.
func (proc *Proc) Def() core.NodeDef { return proc.NodeDef }

// Bounds implements core.Geom.
func (proc *Proc) Bounds(time float32) m.BoundingBox {
	// This is reasonable but should take from the actual given bounds

	return proc.bounds[0]
}

// Trace implements core.Primitive.
func (proc *Proc) Trace(ray *core.Ray, sg *core.ShaderContext) bool {
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
	transform = m.TransformDecompToMatrix4(proc.transformSRT.TimeKey(ray.Time))

	invTransform, _ = m.Matrix4Inverse(transform)

	ray.P = m.Matrix4MulPoint(invTransform, Rp)
	ray.D = m.Matrix4MulVec(invTransform, Rd)
	ray.Setup()

	hit := false

	for _, geom := range proc.Geom {
		if geom.Trace(ray, sg) {
			hit = true

			if ray.Type&core.RayTypeShadow == 0 {
				break
			}
		}
	}

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
func (proc *Proc) PreRender() error {

	proc.handler = lookupHandler(proc.Handler)

	if proc.handler == nil {
		return fmt.Errorf("Proc.Prerender: Unable to find handler %v", proc.Handler)
	}

	// lookup handler
	err := proc.handler.Init(proc, proc.DataString, proc.Userdata)

	if err != nil {
		fmt.Printf("Proc.Prerender: %v\n", err)
	}

	if proc.Transform.Elems == nil {
		proc.Transform.Elems = append(proc.Transform.Elems, m.Matrix4Identity)
		proc.Transform.MotionKeys = 1
	}

	for i := range proc.Transform.Elems {
		proc.transformSRT = append(proc.transformSRT, m.TransformDecompMatrix4(proc.Transform.Elems[i]))
	}

	for i := range proc.BMin.Elems {
		box := m.BoundingBox{}

		box.GrowVec3(proc.BMin.Elems[i])
		box.GrowVec3(proc.BMax.Elems[i])

		proc.bounds = append(proc.bounds, box)
	}

	return nil
}

// PostRender is a core.Node method.
func (proc *Proc) PostRender() error { return nil }

// MotionKeys returns the number of motion keys.
func (proc *Proc) MotionKeys() int {

	return len(proc.transformSRT)

}

func create() (core.Node, error) {
	mfile := Proc{}

	return &mfile, nil
}

var handlers = map[string]func() (Handler, error){}

func lookupHandler(name string) Handler {
	create := handlers[name]

	if create != nil {
		handler, err := create()

		if err != nil {
			return nil
		}

		return handler
	}

	return nil

}

// RegisterHandler is called by handlers to register their creation functions.
func RegisterHandler(name string, create func() (Handler, error)) error {
	handlers[name] = create
	return nil
}

func init() {
	nodes.Register("Proc", create)
}
