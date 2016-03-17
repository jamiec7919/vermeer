// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package material

import (
	"github.com/jamiec7919/vermeer/material/texture"
	m "github.com/jamiec7919/vermeer/math"
)

type Id int32

const ID_NONE Id = -1

type Material struct {
	Name  string
	Sides int
	BSDF  [2]BSDF // bsdf for sides
	//Medium [2]Medium  // medium material
	EDF EDF // Emission distribution for light materials (nil is not a light)

	BumpMap *BumpMap
}

type BumpMap struct {
	Map   MapSampler
	Scale float32
}

func (mtl *Material) ApplyBumpMap(surf *SurfacePoint) {
	delta := float32(1) / float32(6000)
	//log.Printf("Bump %v %v %v %v", mtl.BumpMap.Map.SampleRGB(surf.UV[0][0]-delta, surf.UV[0][1], delta, delta)[0], mtl.BumpMap.Map.SampleRGB(surf.UV[0][0]+delta, surf.UV[0][1], delta, delta)[0], surf.UV[0][0]-delta, surf.UV[0][0]+delta)
	Bu := (1.0 / (2.0 * delta)) * mtl.BumpMap.Scale * (mtl.BumpMap.Map.SampleRGB(surf.UV[0][0]-delta, surf.UV[0][1], delta, delta)[0] - mtl.BumpMap.Map.SampleRGB(surf.UV[0][0]+delta, surf.UV[0][1], delta, delta)[0])
	Bv := (1.0 / (2.0 * delta)) * mtl.BumpMap.Scale * (mtl.BumpMap.Map.SampleRGB(surf.UV[0][0], surf.UV[0][1]-delta, delta, delta)[0] - mtl.BumpMap.Map.SampleRGB(surf.UV[0][0], surf.UV[0][1]+delta, delta, delta)[0])
	//log.Printf("Bump %v %v %v", Bu, Bv, surf.Ns)
	surf.Ns = m.Vec3Add(surf.Ns, m.Vec3Sub(m.Vec3Scale(Bu, m.Vec3Cross(surf.Ns, surf.Pv[0])), m.Vec3Scale(Bv, m.Vec3Cross(surf.Ns, surf.Pu[0]))))
	//log.Printf("%v", surf.Ns)
	surf.SetupTangentSpace(surf.Ns)

}

type ConstantMap struct {
	C [3]float32
}

func (c *ConstantMap) SampleRGB(s, t, ds, dt float32) (out [3]float32) {
	out[0] = c.C[0]
	out[1] = c.C[1]
	out[2] = c.C[2]
	return
}

func (c *ConstantMap) SampleScalar(s, t, ds, dt float32) (out float32) {
	out = c.C[0]
	return
}

type TextureMap struct {
	Filename string
}

func (c *TextureMap) SampleRGB(s, t, ds, dt float32) (out [3]float32) {
	return texture.SampleRGB(c.Filename, s, t, ds, dt)
}

func (c *TextureMap) SampleScalar(s, t, ds, dt float32) (out float32) {
	q := texture.SampleRGB(c.Filename, s, t, ds, dt)
	return q[0]

}
