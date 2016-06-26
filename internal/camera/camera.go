// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package camera

import (
	"fmt"
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
We(y0->y1) = d^2 / (Apixel * cos^3	theta_o),
P(y0->y1) = d^2 / (Apixel * cos^3	theta_o),

cos theta_o = D dot camera.N  (camera.N = -camera.W)
Apixel is pixel density of image with respect to screen size.

For a circular thin lense model camera:
Spatial
We(y0) = 1 / (pi * r^2)
P_A(y0) = 1 / (pi * r^2)

r is radius of lens/apeture
Angular
(same as pinhole)

Should sample position on lens, and directional component.
Note that for bpt these should be separate as might want to join to
lens.  However this should mean that light paths can contribute to
*any* pixel when joining to lens. (and indeed may be splatted if
there is a multi-pixel filter)
*/

// The Camera node represents a 3D camera, it can be specified as a 'lookat' or matrix and
// may have motion keys for its parameters.
type Camera struct {
	NodeName string `node:"Name"`

	Type   string            // 'lookat' 'matrix'
	From   core.PointArray   // eye point
	Target core.PointArray   `node:"To"`
	Roll   core.Float32Array // rotation around z axis
	Up     m.Vec3

	U, V, W m.Vec3 // orthonormal basis

	Aspect float32 // aspect ratio
	Fov    float32 // field of view (X), Y is calculated from aspect
	Focal  float32 // focal length

	WorldToLocal core.MatrixArray
	LocalToWorld core.MatrixArray

	decomp []m.TransformDecomp

	TanThetaFocal float32 // = tan(Fov/2)*Focal

	L, R, T, B float32
	Radius     float32
}

func degToRad(deg float32) float32 { return deg * m.Pi / 180.0 }

// PreRender is a core.Node method.
func (c *Camera) PreRender(rc *core.RenderContext) error {
	if c.Aspect == 0.0 {
		c.Aspect = rc.FrameAspect()
	}

	if c.From.MotionKeys < 2 && c.Target.MotionKeys < 2 {
		c.calcBasisLookat(c.From.Elems[0], c.Target.Elems[0], c.Up, 0)
	}

	c.TanThetaFocal = m.Tan(degToRad(c.Fov/2)) * c.Focal

	c.calcLookatMatrices()

	return nil
}

// PostRender is a core.Node method.
func (c *Camera) PostRender(*core.RenderContext) error { return nil }

// Name is a core.Node method.
func (c *Camera) Name() string { return c.NodeName }

func (c *Camera) calcLookatMatrices() {
	if c.Target.MotionKeys > c.From.MotionKeys {

		for i := 0; i < c.Target.MotionKeys; i++ {
			time := float32(i) / float32(c.Target.MotionKeys)

			k := time * float32(c.From.MotionKeys-1)

			t := k - m.Floor(k)

			key := int(m.Floor(k))
			key2 := int(m.Ceil(k))

			P := m.Vec3Lerp(c.From.Elems[key], c.From.Elems[key2], t)

			W := m.Vec3Normalize(m.Vec3Sub(P, c.Target.Elems[i])) // Note W points away from target
			u := m.Vec3Normalize(m.Vec3Cross(c.Up, W))
			v := m.Vec3Normalize(m.Vec3Cross(u, W))

			roll := float32(0)

			if c.Roll.Elems != nil {
				k := time * float32(c.Roll.MotionKeys-1)

				t := k - m.Floor(k)

				key := int(m.Floor(k))
				key2 := int(m.Ceil(k))

				roll = (1-t)*c.Roll.Elems[key] + t*c.Roll.Elems[key2]
			}

			U := m.Vec3Add(m.Vec3Scale(m.Cos(roll), u), m.Vec3Scale(m.Sin(roll), v))
			V := m.Vec3Add(m.Vec3Scale(-m.Sin(roll), u), m.Vec3Scale(m.Cos(roll), v))

			mtx := m.Matrix4Mul(m.Matrix4Translate(P[0], P[1], P[2]), m.Matrix4Basis(U, V, W))

			c.LocalToWorld.Elems = append(c.LocalToWorld.Elems, mtx)
			c.LocalToWorld.MotionKeys++
		}
	} else {
		for i := 0; i < c.From.MotionKeys; i++ {
			time := float32(i) / float32(c.From.MotionKeys)

			k := time * float32(c.Target.MotionKeys-1)

			t := k - m.Floor(k)

			key := int(m.Floor(k))
			key2 := int(m.Ceil(k))

			P := m.Vec3Lerp(c.Target.Elems[key], c.Target.Elems[key2], t)

			W := m.Vec3Normalize(m.Vec3Sub(c.From.Elems[i], P)) // Note W points away from target
			u := m.Vec3Normalize(m.Vec3Cross(c.Up, W))
			v := m.Vec3Normalize(m.Vec3Cross(u, W))

			roll := float32(0)

			if c.Roll.Elems != nil {
				k := time * float32(c.Roll.MotionKeys-1)

				t := k - m.Floor(k)

				key := int(m.Floor(k))
				key2 := int(m.Ceil(k))

				roll = (1-t)*c.Roll.Elems[key] + t*c.Roll.Elems[key2]
			}

			U := m.Vec3Add(m.Vec3Scale(m.Cos(roll), u), m.Vec3Scale(m.Sin(roll), v))
			V := m.Vec3Add(m.Vec3Scale(-m.Sin(roll), u), m.Vec3Scale(m.Cos(roll), v))

			mtx := m.Matrix4Mul(m.Matrix4Translate(c.From.Elems[i][0], c.From.Elems[i][1], c.From.Elems[i][2]), m.Matrix4Basis(U, V, W))

			c.LocalToWorld.Elems = append(c.LocalToWorld.Elems, mtx)
			c.LocalToWorld.MotionKeys++
		}

	}

	for i := range c.LocalToWorld.Elems {
		c.decomp = append(c.decomp, m.TransformDecompMatrix4(c.LocalToWorld.Elems[i]))
	}
}

func (c *Camera) calcBasisLookat(from, to, up m.Vec3, roll float32) {

	c.W = m.Vec3Normalize(m.Vec3Sub(from, to)) // Note W points away from target
	u := m.Vec3Normalize(m.Vec3Cross(up, c.W))
	v := m.Vec3Normalize(m.Vec3Cross(u, c.W))

	c.U = m.Vec3Add(m.Vec3Scale(m.Cos(roll), u), m.Vec3Scale(m.Sin(roll), v))
	c.V = m.Vec3Add(m.Vec3Scale(-m.Sin(roll), u), m.Vec3Scale(m.Cos(roll), v))

}

/*
	s := vm.Vec3Add(c.Eye, vm.Vec3Sub(vm.Vec3Add(vm.Vec3Scale(u, c.U), vm.Vec3Scale(v, c.V)), vm.Vec3Scale(c.D, c.W)))

	x, y := sample.UniformDisk2D(0.01, r0, r1)
	e := vm.Vec3Add(c.Eye, vm.Vec3Add(vm.Vec3Scale(x, c.U), vm.Vec3Scale(y, c.V)))
	D := vm.Vec3Normalize(vm.Vec3Sub(s, e))

*/
/*
// SampleLensArea samples the lens according to area.
func (c *Camera) SampleLensArea(rnd *rand.Rand, P *m.Vec3, We *float32, pdf *float32) error {
	*P = c.From

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

// SampleImagePlaneDir samples a direction from the image plane.
func (c *Camera) SampleImagePlaneDir(u, v float32, P m.Vec3, rnd *rand.Rand, omegaO *m.Vec3, We *float32, pdf *float32) error {

	camu := u * c.TanThetaFocal
	camv := v * (c.TanThetaFocal / c.Aspect)

	s := m.Vec3Sub(m.Vec3Add(m.Vec3Scale(camu, c.U), m.Vec3Scale(camv, c.V)), m.Vec3Scale(c.Focal, c.W))
	*omegaO = m.Vec3Normalize(m.Vec3Sub(m.Vec3Add(c.From, s), P))

	cosOmegaO := m.Vec3Dot(*omegaO, m.Vec3Neg(c.W))
	Apixel := float32(1.0) // pixel density WRT image size
	*We = (c.Focal * c.Focal) / (Apixel * cosOmegaO * cosOmegaO * cosOmegaO)
	*pdf = (c.Focal * c.Focal) / (Apixel * cosOmegaO * cosOmegaO * cosOmegaO)
	return nil
}
*/

// ComputeRay calculates a position and direction for a sampled ray.
func (c *Camera) ComputeRay(u, v, time float32, rnd *rand.Rand, ray *core.RayData, sg *core.ShaderGlobals) {
	M := m.Matrix4Identity

	if c.decomp != nil {
		k := time * float32(len(c.decomp)-1)

		t := k - m.Floor(k)

		key := int(m.Floor(k))
		key2 := int(m.Ceil(k))

		trn := m.TransformDecompLerp(c.decomp[key], c.decomp[key2], t)

		M = m.TransformDecompToMatrix4(trn)
	}

	// D = || u*U + v*V - d*W  ||

	camu := u * c.TanThetaFocal
	camv := v * (c.TanThetaFocal / c.Aspect)

	U := m.Vec3{1, 0, 0}
	V := m.Vec3{0, 1, 0}
	W := m.Vec3{0, 0, 1}

	s := m.Vec3Sub(m.Vec3Add(m.Vec3Scale(camu, U), m.Vec3Scale(camv, V)), m.Vec3Scale(c.Focal, W))

	D := m.Vec3{0, 0, 1}
	P := m.Vec3{}

	if c.Radius > 0.0 {
		x, y := sample.UniformDisk2D(c.Radius, rnd.Float32(), rnd.Float32())
		e := m.Vec3Add(m.Vec3Scale(x, U), m.Vec3Scale(y, V))
		D = m.Matrix4MulVec(M, m.Vec3Normalize(m.Vec3Sub(s, e)))
		P = m.Matrix4MulPoint(M, e)
	} else {
		D = m.Matrix4MulVec(M, m.Vec3Normalize(s))
		P = m.Matrix4MulPoint(M, m.Vec3{})
	}
	//D = D

	ray.Init(core.RayCamera, P, D, m.Inf(1), sg)
	//	log.Printf("%v %v %v %v", D, u, v, vm.Vec3Add(vm.Vec3Scale(u, c.U), vm.Vec3Scale(v, c.V)))
	return
}

var cameraCount = 0

func init() {
	nodes.Register("Camera", func() (core.Node, error) {
		cam := Camera{Focal: 12, Fov: 90, NodeName: fmt.Sprintf("camera<%v>", cameraCount)}

		cameraCount++

		return &cam, nil
	})
}
