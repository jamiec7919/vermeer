// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/cheggaaa/pb"
	"github.com/jamiec7919/vermeer/internal/nodeparser"
	"github.com/jamiec7919/vermeer/material"
	m "github.com/jamiec7919/vermeer/math"
	"log"
	"math/rand"
	"sync"
	"time"
)

const TILESIZE = 64
const MAXGOROUTINES = 5
const NSAMP = 16

type Frame struct {
	w, h   int
	du, dv float32
	camera Camera
	scene  *Scene
	rc     *RenderContext
	bar    *pb.ProgressBar
}

func (f *Frame) Aspect() float32 { return float32(f.w) / float32(f.h) }

type Scene struct {
	prims  []Primitive
	lights []Light
}

func (s *Scene) VisRay(ray *RayData) {
	// d is not normalized and Tclosest == 1-VisRayEpsilon
	for _, prim := range s.prims {
		prim.VisRay(ray)

		if !ray.IsVis() { // Early out if we have a blocker
			return
		}
	}
}

func (s *Scene) TraceRay(ray *RayData) {
	for _, prim := range s.prims {
		prim.TraceRay(ray)
	}

}

type PreviewFrame struct {
	W, H int
	Buf  []uint8
}

type RenderContext struct {
	globals   Globals
	imgbuf    []float32
	frames    []Frame
	nodes     []Node
	scene     Scene
	cameras   []Camera
	materials []*material.Material

	PreviewChan chan PreviewFrame
}

func (rc *RenderContext) GetMaterial(id material.Id) *material.Material {
	if rc.materials != nil && id != material.ID_NONE && int(id) < len(rc.materials) {
		return rc.materials[int(id)]
	}
	return nil
}

func NewRenderContext() *RenderContext {
	rc := &RenderContext{}
	rc.globals.XRes = 256
	rc.globals.YRes = 256
	rc.globals.MaxGoRoutines = MAXGOROUTINES
	return rc
}

func (rc *RenderContext) PreRender() error {
	// pre and fixup nodes
	for _, node := range rc.nodes {
		if err := node.PreRender(rc); err != nil {
			return err
		}
	}

	return nil
}

type WorkItem struct {
	x, y, w, h int
	samples    []float32
}

/* This should return an rgb sample to be accumulated for the pixel */
func samplePixel(x, y int, frame *Frame, rnd *rand.Rand, ray *RayData) (r, g, b float32) {
	//log.Printf("Pix %v %v", x, y)
	r0 := rnd.Float32()
	r1 := rnd.Float32()

	u := (float32(x) + r0) * frame.du
	v := (float32(y) + r1) * frame.dv

	lambda := ((720 - 450) * rnd.Float32()) + 450

	P, D := frame.camera.ComputeRay(-1+u, 1-v, rnd)
	fullsample := material.Spectrum{Lambda: lambda}
	contrib := material.Spectrum{Lambda: fullsample.Lambda}
	contrib.FromRGB(1, 1, 1)

	for depth := 0; depth < 4; depth++ {

		ray.InitRay(P, D)

		frame.scene.TraceRay(ray)

		var surf material.SurfacePoint

		if ray.GetHitSurface(&surf) == nil {

			mtl := frame.rc.GetMaterial(material.Id(surf.MtlId))

			if mtl == nil { // can't do much with no material
				return
			}

			Vout := m.Vec3Neg(D)
			//			d := m.Vec3Dot(surf.N, Vout)

			//if d < 0.0 { // backface hit
			//	return
			//}

			//surf.Ns = surf.WorldToTangent(m.Vec3Normalize(surf.Ns))

			//if m.Vec3Dot(surf.Ns, surf.N) < 0 {
			//		Ns := vm.Vec3Add(shade.Ns, vm.Vec3Scale(2*vm.Vec3Dot(shade.Ns, shade.Ng), shade.Ng))

			//	surf.Ns = m.Vec3Neg(surf.Ns) // Should mirror in Ng really instead of -ve?
			//}

			omega_i := surf.WorldToTangent(Vout)

			bsdf := mtl.BSDF[0]

			if bsdf == nil { // can't do much without BSDF
				//return
			}

			//var samp_pdf float64
			//var omega_o m.Vec3

			// Assume that no transmission, so offset surface point out from surface
			surf.OffsetP(1)

			if !bsdf.IsDelta(&surf) {

				if len(frame.scene.lights) > 0 {
					nls := 4
					lightsamples := 0
					if depth > 0 {
						nls = 1
					}
					for i := 0; i < nls; i++ {
						var P material.SurfacePoint
						var pdf float64

						if frame.scene.lights[0].SampleArea(&surf, rnd, &P, &pdf) == nil {
							V := m.Vec3Sub(P.P, surf.P)

							if m.Vec3Dot(V, surf.Ns) > 0.0 && m.Vec3Dot(V, P.N) < 0.0 {
								ray.InitVisRay(surf.P, P.P)
								frame.scene.VisRay(ray)
								if ray.IsVis() {
									lightsamples++
									Vnorm := m.Vec3Normalize(V)

									lightm := frame.rc.GetMaterial(material.Id(P.MtlId))
									Le := material.Spectrum{Lambda: contrib.Lambda}
									lightm.EDF.Eval(&P, P.WorldToTangent(m.Vec3Neg(Vnorm)), &Le)

									rho := material.Spectrum{Lambda: contrib.Lambda}

									bsdf.Eval(&surf, omega_i, surf.WorldToTangent(Vnorm), &rho)
									geom := m.Abs(m.Vec3Dot(Vnorm, surf.Ns)) * m.Abs(m.Vec3Dot(Vnorm, P.N)) / m.Vec3Length2(V)
									Le.Mul(rho)
									Le.Mul(contrib)
									Le.Scale(geom / (float32(pdf) * float32(nls)))

									fullsample.Add(Le)
									//log.Printf("contrib:", contrib)
								}
							}
						}
					}
				}
			}

			var omega_o m.Vec3
			var pdf float64
			count := 0
		resample:
			rho := material.Spectrum{Lambda: fullsample.Lambda}
			bsdf.Sample(&surf, omega_i, rnd, &omega_o, &rho, &pdf)

			D = surf.TangentToWorld(omega_o)

			if m.Vec3Dot(D, surf.N) < 0 {
				if count < 5 {
					goto resample
				} else {
					log.Printf("Exceeded 5 resample")
				}
			}

			contrib.Mul(rho)
			contrib.Scale(omega_o[2] / float32(pdf))

			P = surf.P
			//log.Printf("%v %v", x, y)
			//return contrib.ToRGB()
			//r = m.Vec3Dot(surf.N, m.Vec3Neg(D))
			//g = m.Vec3Dot(surf.N, m.Vec3Neg(D))
			//b = m.Vec3Dot(surf.N, m.Vec3Neg(D))
		} else { // Escaped scene
			break
		}
	}
	return fullsample.ToRGB()
}

// NOTE: we return the raydata here even though it is ignored in order to ensure that ray is
// heap allocated (for alignment purposes)
func renderFunc(n int, frame *Frame, c chan *WorkItem, done chan *WorkItem, wg *sync.WaitGroup) *RayData {
	defer wg.Done()
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	ray := &RayData{}
	for w := range c {
		for j := 0; j < w.h; j++ {
			for i := 0; i < w.w; i++ {
				r, g, b := samplePixel(i+w.x, j+w.y, frame, rnd, ray)

				w.samples[((i+w.x)+(j+w.y)*frame.w)*3+0] = (w.samples[((i+w.x)+(j+w.y)*frame.w)*3+0]*float32(n) + m.Clamp(r*1000, 0, 255)) / float32(n+1)
				w.samples[((i+w.x)+(j+w.y)*frame.w)*3+1] = (w.samples[((i+w.x)+(j+w.y)*frame.w)*3+1]*float32(n) + m.Clamp(g*1000, 0, 255)) / float32(n+1)
				w.samples[((i+w.x)+(j+w.y)*frame.w)*3+2] = (w.samples[((i+w.x)+(j+w.y)*frame.w)*3+2]*float32(n) + m.Clamp(b*1000, 0, 255)) / float32(n+1)

				if frame.bar != nil {
					frame.bar.Increment()
				}
			}
		}

		done <- w
	}

	return ray
}

func tonemap(w, h int, hdr_rgb []float32, buf []uint8) {
	// Tone map into buffer
	for i := 0; i < w*h*3; i += 3 {
		buf[i] = uint8(hdr_rgb[i])
		buf[i+1] = uint8(hdr_rgb[i+1])
		buf[i+2] = uint8(hdr_rgb[i+2])

	}

}

func (rc *RenderContext) FrameAspect() float32 {
	return float32(rc.globals.XRes) / float32(rc.globals.YRes)
}

func (rc *RenderContext) Render(finish chan bool) error {
	// render frames as given in frames (could be progressive)
	var frame Frame

	if len(rc.frames) == 0 {
		if node := rc.FindNode("camera"); node != nil {
			frame.camera = node.(Camera)
		}
	}

	if frame.camera == nil {
		return ErrNoCamera
	}

	frame.rc = rc
	frame.scene = &rc.scene
	frame.w = rc.globals.XRes
	frame.h = rc.globals.YRes
	frame.du = 2.0 / float32(frame.w)
	frame.dv = 2.0 / float32(frame.h)

	if rc.globals.UseProgress {
		frame.bar = pb.StartNew(rc.globals.XRes * rc.globals.YRes)
	}

	buf := make([]float32, frame.w*frame.h*3)
L:
	for k := 0; true; k++ {

		var wg sync.WaitGroup
		workChan := make(chan *WorkItem)
		done := make(chan *WorkItem)

		for n := 0; n < rc.globals.MaxGoRoutines; n++ {
			wg.Add(1)
			go renderFunc(k, &frame, workChan, done, &wg)
		}

		complete := make(chan []float32)
		go func() {
			var q []*WorkItem
			for d := range done {
				q = append(q, d)
			}
			/*
				for k := range q {
					for j := 0; j < q[k].h; j++ {
						for i := 0; i < q[k].w; i++ {
							buf[((i+q[k].x)+(j+q[k].y)*frame.w)*3+0] = q[k].samples[(i+(j*q[k].w))*3+0]
							buf[((i+q[k].x)+(j+q[k].y)*frame.w)*3+1] = q[k].samples[(i+(j*q[k].w))*3+1]
							buf[((i+q[k].x)+(j+q[k].y)*frame.w)*3+2] = q[k].samples[(i+(j*q[k].w))*3+2]
						}
					}
				}
			*/
			complete <- buf
		}()

		for j := 0; j < frame.h; j += TILESIZE {
			for i := 0; i < frame.w; i += TILESIZE {

				workChan <- &WorkItem{x: i, y: j, w: TILESIZE, h: TILESIZE, samples: buf /* make([]float32, TILESIZE*TILESIZE*3)*/}
			}
		}

		close(workChan)
		wg.Wait()
		close(done)

		rc.imgbuf = <-complete

		if rc.PreviewChan != nil {
			fr := PreviewFrame{
				W:   rc.globals.XRes,
				H:   rc.globals.YRes,
				Buf: make([]uint8, 3*rc.globals.XRes*rc.globals.YRes),
			}

			tonemap(rc.globals.XRes, rc.globals.YRes, rc.imgbuf, fr.Buf)

			rc.PreviewChan <- fr
		}

		select {
		case <-finish:
			break L
		default:
		}
	}

	if frame.bar != nil {
		frame.bar.FinishPrint("Render Complete")
	}

	return nil
}

func (rc *RenderContext) PostRender() error {
	// post process image
	for _, node := range rc.nodes {
		if err := node.PostRender(rc); err != nil {
			return err
		}
	}

	return nil
}

func (rc *RenderContext) GetMaterialId(name string) material.Id {
	for id, mtl := range rc.materials {
		if mtl.Name == name {
			return material.Id(id)
		}
	}

	return material.ID_NONE
}

func (rc *RenderContext) AddMaterial(name string, mtl *material.Material) material.Id {
	mtl.Name = name

	id := len(rc.materials)

	rc.materials = append(rc.materials, mtl)

	return material.Id(id)
}

func (rc *RenderContext) AddNode(node Node) {
	rc.nodes = append(rc.nodes, node)

	switch t := node.(type) {
	case Camera:
		rc.cameras = append(rc.cameras, t)
	case Primitive:
		rc.scene.prims = append(rc.scene.prims, t)
	case Light:
		rc.scene.lights = append(rc.scene.lights, t)
	case Material:
		rc.AddMaterial(t.Name(), t.Material())
	}
}

func (rc *RenderContext) FindNode(name string) Node {
	for _, node := range rc.nodes {
		if node.Name() == name {
			return node
		}
	}
	return nil
}

func (rc *RenderContext) LoadNodeFile(filename string) error {
	return nodeparser.Parse(rc, filename)
}

func (rc *RenderContext) DispatchNode(_node interface{}) error {
	node, ok := _node.(Node)

	if !ok {
		nodes, ok := _node.([]Node)

		if ok {
			for _, n := range nodes {
				rc.AddNode(n)
			}
			return nil
		}
		return ErrNotNode
	}
	rc.AddNode(node)
	return nil
}

func (rc *RenderContext) CreateObj(method string, _params map[string]interface{}) (interface{}, error) {
	params := Params(_params)

	create, present := objTypes[method]

	if present {
		return create(rc, params)
	}

	return nil, ErrNodeNotRegistered
}

func (rc *RenderContext) Error(err error) error {
	log.Printf("Parse error: %v", err)
	return nil
}

type Node interface {
	Name() string
	PreRender(*RenderContext) error
	PostRender(*RenderContext) error
}

var objTypes = map[string]func(*RenderContext, Params) (interface{}, error){}

func RegisterType(name string, create func(*RenderContext, Params) (interface{}, error)) error {
	if objTypes[name] == nil {
		objTypes[name] = create
		return nil
	}
	return ErrNodeAlreadyRegistered
}
