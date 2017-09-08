// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package camera

import (
	"fmt"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/math/sample"
	//"math/rand"
	"github.com/jamiec7919/vermeer/core"
	param "github.com/jamiec7919/vermeer/core/param"
	"github.com/jamiec7919/vermeer/nodes"
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
	NodeDef  core.NodeDef `node:"-"`
	NodeName string       `node:"Name"`

	Type   string             // 'lookat' 'matrix'
	From   param.PointArray   `node:",opt"` // eye point
	Target param.PointArray   `node:"To,opt"`
	Roll   param.Float32Array `node:",opt"` // rotation around z axis
	Up     m.Vec3             `node:",opt"`

	U, V, W m.Vec3 `node:"-"` // orthonormal basis

	Aspect float32 `node:",opt"` // aspect ratio
	Fov    float32 `node:",opt"` // field of view (X), Y is calculated from aspect
	Focal  float32 `node:",opt"` // focal length

	WorldToLocal param.MatrixArray `node:",opt"`
	LocalToWorld param.MatrixArray `node:",opt"`

	decomp []m.Transform

	TanThetaFocal float32 `node:"-"` // = tan(Fov/2)*Focal

	L, R, T, B float32 `node:",opt"`
	Radius     float32 `node:",opt"`
}

var _ core.Node = (*Camera)(nil)

func degToRad(deg float32) float32 { return deg * m.Pi / 180.0 }

// PreRender is a core.Node method.
func (c *Camera) PreRender() error {
	if c.Aspect == 0.0 {
		c.Aspect = core.FrameAspect()
	}

	//if c.From.MotionKeys < 2 && c.Target.MotionKeys < 2 {
	//	c.calcBasisLookat(c.From.Elems[0], c.Target.Elems[0], c.Up, 0)
	//}

	c.TanThetaFocal = m.Tan(degToRad(c.Fov/2)) * c.Focal

	if c.Type == "LookAt" {
		c.calcLookatMatrices()
	} else {
		c.matrixCalc()
	}

	return nil
}

// PostRender is a core.Node method.
func (c *Camera) PostRender() error { return nil }

// Name is a core.Node method.
func (c *Camera) Name() string { return c.NodeName }

// Def is a core.Node method.
func (c *Camera) Def() core.NodeDef { return c.NodeDef }

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

			mtx := m.Matrix4Mul(m.Matrix4TransformTranslate(P.X, P.Y, P.Z), m.Matrix4Basis(U, V, W))

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

			mtx := m.Matrix4Mul(m.Matrix4TransformTranslate(c.From.Elems[i].X, c.From.Elems[i].Y, c.From.Elems[i].Z), m.Matrix4Basis(U, V, W))

			c.LocalToWorld.Elems = append(c.LocalToWorld.Elems, mtx)
			c.LocalToWorld.MotionKeys++
		}

	}

	for i := range c.LocalToWorld.Elems {

		c.decomp = append(c.decomp, m.Matrix4ToTransform(c.LocalToWorld.Elems[i]))
	}
}

func (c *Camera) calcBasisLookat(from, to, up m.Vec3, roll float32) {

	c.W = m.Vec3Normalize(m.Vec3Sub(from, to)) // Note W points away from target
	u := m.Vec3Normalize(m.Vec3Cross(up, c.W))
	v := m.Vec3Normalize(m.Vec3Cross(u, c.W))

	c.U = m.Vec3Add(m.Vec3Scale(m.Cos(roll), u), m.Vec3Scale(m.Sin(roll), v))
	c.V = m.Vec3Add(m.Vec3Scale(-m.Sin(roll), u), m.Vec3Scale(m.Cos(roll), v))

}

func (c *Camera) matrixCalc() {
	for _, mtx := range c.WorldToLocal.Elems {
		srt, _ := m.Matrix4Inverse(mtx)
		c.LocalToWorld.Elems = append(c.LocalToWorld.Elems, srt)
		c.LocalToWorld.MotionKeys++
	}

	for i := range c.LocalToWorld.Elems {
		c.decomp = append(c.decomp, m.Matrix4ToTransform(c.LocalToWorld.Elems[i]))
	}

}

// ComputeRay calculates a position and direction for a sampled ray.
// x,y are the raster position, lensU,lensV are in [0,1)x[0,1)
func (c *Camera) ComputeRay(sc *core.ShaderContext, lensU, lensV float64, ray *core.Ray) {

	M := m.Matrix4Identity()

	if c.decomp != nil {
		k := sc.Time * float32(len(c.decomp)-1)

		t := k - m.Floor(k)

		key := int(m.Floor(k))
		key2 := int(m.Ceil(k))

		trn := m.TransformLerp(c.decomp[key], c.decomp[key2], t)

		M = m.TransformToMatrix4(trn)
	}

	// D = || u*U + v*V - d*W  ||
	ray.X = sc.X
	ray.Y = sc.Y
	ray.Sx = sc.Sx
	ray.Sy = sc.Sy

	camu := float32(ray.Sx) * float32(c.TanThetaFocal)
	camv := float32(ray.Sy) * float32(c.TanThetaFocal/core.FrameAspect())

	U := m.Vec3{1, 0, 0}
	V := m.Vec3{0, 1, 0}
	W := m.Vec3{0, 0, 1}

	s := m.Vec3Sub(m.Vec3Add(m.Vec3Scale(camu, U), m.Vec3Scale(camv, V)), m.Vec3Scale(c.Focal, W))

	D := m.Vec3{0, 0, 1}
	P := m.Vec3{}

	dx := float32(0)
	dy := float32(0)

	var d m.Vec3
	w, h := core.FrameMetrics()

	if c.Radius > 0.0 {

		x, y := sample.UniformDisk2D(c.Radius, float32(lensU), float32(lensV))
		e := m.Vec3Add(m.Vec3Scale(x, U), m.Vec3Scale(y, V))
		//D = m.Matrix4MulVec(M, m.Vec3Normalize(m.Vec3Sub(s, e)))
		d = m.Matrix4MulVec(M, m.Vec3Sub(s, e))

		camu2 := float32(ray.Sx+(2/float32(w))) * float32(c.TanThetaFocal)
		camv2 := float32(ray.Sy+(2/float32(h))) * float32(c.TanThetaFocal/c.Aspect)
		sx := m.Vec3Sub(m.Vec3Add(m.Vec3Scale(camu2, U), m.Vec3Scale(camv, V)), m.Vec3Scale(c.Focal, W))
		sy := m.Vec3Sub(m.Vec3Add(m.Vec3Scale(camu, U), m.Vec3Scale(camv2, V)), m.Vec3Scale(c.Focal, W))
		d2x := m.Matrix4MulVec(M, m.Vec3Sub(sx, e))
		d2y := m.Matrix4MulVec(M, m.Vec3Sub(sy, e))
		dx = m.Vec3Length(m.Vec3Sub(d2x, d))
		dy = m.Vec3Length(m.Vec3Sub(d2y, d))

		D = m.Vec3Normalize(d)
		P = m.Matrix4MulPoint(M, e)
	} else {
		//D = m.Matrix4MulVec(M, m.Vec3Normalize(s))
		d = m.Matrix4MulVec(M, s)
		D = m.Vec3Normalize(d)
		P = m.Matrix4MulPoint(M, m.Vec3{})

		camu2 := float32(ray.Sx+(2/float32(w))) * float32(c.TanThetaFocal)
		camv2 := float32(ray.Sy+(2/float32(h))) * float32(c.TanThetaFocal/c.Aspect)
		sx := m.Vec3Sub(m.Vec3Add(m.Vec3Scale(camu2, U), m.Vec3Scale(camv, V)), m.Vec3Scale(c.Focal, W))
		sy := m.Vec3Sub(m.Vec3Add(m.Vec3Scale(camu, U), m.Vec3Scale(camv2, V)), m.Vec3Scale(c.Focal, W))
		d2x := m.Matrix4MulVec(M, m.Vec3Sub(sx, m.Vec3{}))
		d2y := m.Matrix4MulVec(M, m.Vec3Sub(sy, m.Vec3{}))
		dx = m.Vec3Length(m.Vec3Sub(d2x, d))
		dy = m.Vec3Length(m.Vec3Sub(d2y, d))

	}
	//D = D

	// Right and Up vectors take into account the world space distance between pixels.
	// A differential of size 1 should then be 1 screen pixel.
	right := m.Vec3Scale(2*c.TanThetaFocal/float32(w), m.Vec3{M[0], M[1], M[2]})
	up := m.Vec3Scale(2*c.TanThetaFocal/(c.Aspect*float32(h)), m.Vec3{M[4], M[5], M[6]})
	//right := m.Vec3{M[0], M[4], M[8]}
	//up := m.Vec3{M[1], M[5], M[9]}

	ray.DdPdx = m.Vec3{}
	ray.DdPdy = m.Vec3{}
	ray.DdDdx = m.Vec3Scale(1/(m.Vec3Dot(d, d)*m.Sqrt(m.Vec3Dot(d, d))), m.Vec3Sub(m.Vec3Scale(m.Vec3Dot(d, d), right), m.Vec3Scale(m.Vec3Dot(d, right), d)))
	ray.DdDdy = m.Vec3Scale(1/(m.Vec3Dot(d, d)*m.Sqrt(m.Vec3Dot(d, d))), m.Vec3Sub(m.Vec3Scale(m.Vec3Dot(d, d), up), m.Vec3Scale(m.Vec3Dot(d, up), d)))

	//fmt.Printf("%v %v %v %v %v\n", ray.DdDdx, ray.DdDdy, m.Vec3Length(ray.DdDdx), m.Vec3Length(ray.DdDdy), m.Vec3Length(m.Vec3Scale(c.Focal, ray.DdDdx)))
	// Should calculate these from world space dimensions
	//w, h := core.FrameMetrics()
	dx = 2.0 / float32(w)
	dy = 2.0 / float32(h)

	_ = dx
	_ = dy

	// Not sure of correct scaling factors here.. or why I added the 0.1 factor, breaks the texturing!
	//sc.Image.PixelDelta[0] = 0.1 * 2 * c.TanThetaFocal / float32(w)
	//sc.Image.PixelDelta[1] = 0.1 * 2 * c.TanThetaFocal / (c.Aspect * float32(h))
	sc.Image.PixelDelta[0] = 1 //2 * c.TanThetaFocal / float32(w)
	sc.Image.PixelDelta[1] = 1 //2 * c.TanThetaFocal / (c.Aspect * float32(h))

	ray.Init(core.RayTypeCamera, P, D, m.Inf(1), 0, sc)

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
