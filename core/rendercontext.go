// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/cheggaaa/pb"
	// "github.com/jamiec7919/vermeer/material"
	m "github.com/jamiec7919/vermeer/math"
	"log"
	"math/rand"
	"sync"
	"time"
)

const TILESIZE = 64
const MAXGOROUTINES = 5
const NSAMP = 16

var rayCount int

type RenderFuncStats struct {
	RayCount, ShadowRayCount int
}

type Frame struct {
	w, h   int
	du, dv float32
	camera Camera
	scene  *Scene
	rc     *RenderContext
	bar    *pb.ProgressBar
}

// Preview windows should implement this
type PreviewWindow interface {
	UpdateFrame(frame PreviewFrame)
	Close()
}

func (f *Frame) Aspect() float32 { return float32(f.w) / float32(f.h) }

type PreviewFrame struct {
	W, H int
	Buf  []uint8
}

func (rc *RenderContext) OutputRes() (int, int) {
	return rc.globals.XRes, rc.globals.YRes
}
func (rc *RenderContext) Image() []float32 {
	return rc.imgbuf
}

type RenderContext struct {
	globals   Globals
	imgbuf    []float32
	frames    []Frame
	nodes     []Node
	scene     Scene
	cameras   []Camera
	materials []Material

	PreviewChan chan PreviewFrame
	preview     PreviewWindow
	finish      chan bool
}

func (rc *RenderContext) GetMaterial(id int32) Material {
	if rc.materials != nil && id != -1 && int(id) < len(rc.materials) {
		return rc.materials[int(id)]
	}
	return nil
}

func NewRenderContext() *RenderContext {
	rc := &RenderContext{}
	rc.globals.XRes = 256
	rc.globals.YRes = 256
	rc.globals.MaxGoRoutines = MAXGOROUTINES
	rc.finish = make(chan bool, 1)
	grc = rc
	return rc
}

func (rc *RenderContext) StartPreview(preview PreviewWindow) error {
	rc.preview = preview
	return nil
}

func (rc *RenderContext) Finish() {
	rc.finish <- true
}

func (rc *RenderContext) PreRender() error {
	// pre and fixup nodes
	// Note that nodes in PreRender may add new nodes, so we must backup and
	// keep track of the existing set so they are only processed once.

	var allnodes []Node

	for rc.nodes != nil {

		nodes := rc.nodes
		rc.nodes = nil
		allnodes = append(allnodes, nodes...)

		for _, node := range nodes {
			if err := node.PreRender(rc); err != nil {
				return err
			}
		}
	}

	rc.nodes = allnodes

	return rc.scene.initAccel()
}

type WorkItem struct {
	x, y, w, h int
	samples    []float32
}

/* This should return an rgb sample to be accumulated for the pixel */
func samplePixel(x, y int, frame *Frame, rnd *rand.Rand, ray *RayData, stats *RenderFuncStats) (r, g, b float32) {
	/*
	  .. Trace AA_count rays around pixel, for each ray that hits different surface/triangle
	    shade that and weight accordingly.

	    Need to get the primitive & face id out of ray intersection. Time needs consideration
	*/

	//log.Printf("Pix %v %v", x, y)
	r0 := rnd.Float32()
	r1 := rnd.Float32()

	u := (float32(x) + r0) * frame.du
	v := (float32(y) + r1) * frame.dv

	lambda := (float32(720-450) * rnd.Float32()) + 450

	P, D := frame.camera.ComputeRay(-1+u, 1-v, rnd)

	sg := &ShaderGlobals{
		Lambda: lambda,
		rnd:    rnd,
	}

	ray.Init(RAY_CAMERA, P, D, m.Inf(1), sg) // Maybe ray types
	stats.RayCount++
	var samp ScreenSample

	if Trace(ray, &samp) {
		return samp.Colour[0], samp.Colour[1], samp.Colour[2]
	}

	return
}

// NOTE: we return the raydata here even though it is ignored in order to ensure that ray is
// heap allocated (for alignment purposes)
func renderFunc(n int, frame *Frame, c chan *WorkItem, done chan *WorkItem, wg *sync.WaitGroup, stats *RenderFuncStats) *RayData {
	defer wg.Done()
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	ray := &RayData{}
	for w := range c {
		for j := 0; j < w.h; j++ {
			for i := 0; i < w.w; i++ {
				r, g, b := samplePixel(i+w.x, j+w.y, frame, rnd, ray, stats)

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

func (rc *RenderContext) Render(maxIter int) error {
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

	startTime := time.Now()
	stats := make([]RenderFuncStats, rc.globals.MaxGoRoutines)

L:
	for k := 0; true; k++ {

		if maxIter >= 0 && k >= maxIter-1 {
			rc.Finish()
		}

		var wg sync.WaitGroup
		workChan := make(chan *WorkItem)
		done := make(chan *WorkItem)

		for n := 0; n < rc.globals.MaxGoRoutines; n++ {
			wg.Add(1)
			go renderFunc(k, &frame, workChan, done, &wg, &stats[n])
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

		if rc.preview != nil {
			fr := PreviewFrame{
				W:   rc.globals.XRes,
				H:   rc.globals.YRes,
				Buf: make([]uint8, 3*rc.globals.XRes*rc.globals.YRes),
			}

			tonemap(rc.globals.XRes, rc.globals.YRes, rc.imgbuf, fr.Buf)

			rc.preview.UpdateFrame(fr)
		}

		select {
		case <-rc.finish:
			if rc.preview != nil {
				rc.preview.Close()
			}
			duration := time.Since(startTime)
			totalRays := 0
			_shadowRays := 0

			for i := range stats {
				totalRays += stats[i].RayCount
				_shadowRays += stats[i].ShadowRayCount
			}
			totalRays = rayCount
			log.Printf("%v iterations, %v (%v rays, %v shadow) %v %v Mr/sec", k+1, duration, totalRays, shadowRays, rayCount, float64(totalRays+shadowRays)/(1000000.0*duration.Seconds()))
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

func (rc *RenderContext) GetMaterialId(name string) int32 {
	for id, mtl := range rc.materials {
		if mtl.Name() == name {
			return int32(id)
		}
	}

	return -1
}

func (rc *RenderContext) addMaterial(mtl Material) {

	id := len(rc.materials)

	rc.materials = append(rc.materials, mtl)

	mtl.SetId(int32(id))
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
		rc.addMaterial(t)
	case *Globals:
		rc.globals = *t
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

func (rc *RenderContext) Error(err error) error {
	log.Printf("Parse error: %v", err)
	return nil
}

type Node interface {
	Name() string
	PreRender(*RenderContext) error
	PostRender(*RenderContext) error
}
