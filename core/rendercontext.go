// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/cheggaaa/pb"
	// "github.com/jamiec7919/vermeer/material"
	"fmt"
	m "github.com/jamiec7919/vermeer/math"
	"log"
	"math/rand"
	"sync"
	"time"
)

// TILESIZE is the size (width and height) in pixels of a render tile.
const TILESIZE = 64

// MAXGOROUTINES is the default number of goroutines to run for rendering.
const MAXGOROUTINES = 5

// NSAMP is the number of samples to take (not used).
// Deprecated: progressive rendering takes as many samples as needed or use maxiter.
const NSAMP = 16

var rayCount uint64
var shadowRays uint64

// Stats collects stats for the whole render.
type Stats struct {
	Duration                 time.Duration
	RayCount, ShadowRayCount uint64
}

func (s *Stats) String() string {
	return fmt.Sprintf("%v	%v	%v/%v", s.Duration, float64(s.RayCount)/(1000000.0*s.Duration.Seconds()), s.RayCount, s.ShadowRayCount)
}

// Frame represents a single frame.
//
// Deprecated: not needed.
type Frame struct {
	w, h   int
	du, dv float32
	camera Camera
	scene  *Scene
	rc     *RenderContext
	bar    *pb.ProgressBar
}

// PreviewWindow is an interface that preview windows should implement.
type PreviewWindow interface {
	UpdateFrame(frame PreviewFrame)
	Close()
}

// Aspect returns the aspect ratio W/H.
func (f *Frame) Aspect() float32 { return float32(f.w) / float32(f.h) }

// PreviewFrame is a buffer passed to the PreviewWindow that has been tonemapped to 24bit RGB.
type PreviewFrame struct {
	W, H int
	Buf  []uint8
}

// OutputRes returns the output resolution.
func (rc *RenderContext) OutputRes() (int, int) {
	return rc.globals.XRes, rc.globals.YRes
}

// Image returns a float32 RGB slice of pixels.
func (rc *RenderContext) Image() []float32 {
	return rc.imgbuf
}

// RenderContext represents everything in the current core API instance.
//
// Deprecated: will only ever be one of these so promote everything to top level and avoid
// passing around the extra pointer.
type RenderContext struct {
	globals   Globals
	imgbuf    []float32
	frames    []Frame
	nodes     []Node
	nodeMap   map[string]Node
	scene     Scene
	cameras   []Camera
	materials []Material

	PreviewChan chan PreviewFrame
	preview     PreviewWindow
	finish      chan bool
}

// GetMaterial returns the shader for the given id.
func (rc *RenderContext) GetMaterial(id int32) Material {
	if rc.materials != nil && id != -1 && int(id) < len(rc.materials) {
		return rc.materials[int(id)]
	}
	return nil
}

// NewRenderContext returns a new RenderContext set to defaults.
//
// Deprecated: Core API is going to be changed to not need a render context as only one frame at
// a time is to be rendered anyway.
func NewRenderContext() *RenderContext {
	rc := &RenderContext{}
	rc.globals.XRes = 256
	rc.globals.YRes = 256
	rc.globals.MaxGoRoutines = MAXGOROUTINES
	rc.finish = make(chan bool, 1)
	rc.nodeMap = make(map[string]Node)
	grc = rc
	return rc
}

// StartPreview is called to initialize the preview window.
func (rc *RenderContext) StartPreview(preview PreviewWindow) error {
	rc.preview = preview
	return nil
}

// Finish is called to notify all listeners that all rendering should finish and exit.
func (rc *RenderContext) Finish() {
	rc.finish <- true
}

// PreRender is called after all nodes are loaded and calls PreRender on all nodes.
// Nodes may add new nodes so PreRender iterates until no new nodes are created.
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

// WorkItem represents a screen tile (note: shouldn't be public).
type WorkItem struct {
	x, y, w, h int
	samples    []float32
}

/* This should return an rgb sample to be accumulated for the pixel */
func samplePixel(x, y int, frame *Frame, rnd *rand.Rand, ray *RayData) (r, g, b float32) {
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
	time := rnd.Float32()

	sg := &ShaderGlobals{
		Lambda: lambda,
		Time:   time,
		rnd:    rnd,
	}

	frame.camera.ComputeRay(-1+u, 1-v, time, rnd, ray, sg)

	var samp ScreenSample

	if Trace(ray, &samp) {
		return samp.Colour[0], samp.Colour[1], samp.Colour[2]
	}

	return
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

func tonemap(w, h int, hdrRGB []float32, buf []uint8) {
	// Tone map into buffer
	for i := 0; i < w*h*3; i += 3 {
		buf[i] = uint8(hdrRGB[i])
		buf[i+1] = uint8(hdrRGB[i+1])
		buf[i+2] = uint8(hdrRGB[i+2])

	}

}

// FrameAspect returns the aspect ration of the global frame size (W/H).
func (rc *RenderContext) FrameAspect() float32 {
	return float32(rc.globals.XRes) / float32(rc.globals.YRes)
}

// Render is called to begin the render process. If maxIter >= 0 only that many iterations
// will be performed before exiting.
func (rc *RenderContext) Render(maxIter int) (stats Stats, err error) {
	// render frames as given in frames (could be progressive)
	var frame Frame

	if len(rc.frames) == 0 {
		if node := rc.FindNode("camera"); node != nil {
			frame.camera = node.(Camera)
		}
	}

	if frame.camera == nil {
		return stats, ErrNoCamera
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

			stats.Duration = duration
			stats.RayCount = rayCount
			stats.ShadowRayCount = shadowRays
			log.Printf("%v iterations, %v (%v rays, %v shadow) %v Mr/sec", k+1, duration, rayCount, shadowRays, float64(rayCount)/(1000000.0*duration.Seconds()))
			break L
		default:
		}
	}

	if frame.bar != nil {
		frame.bar.FinishPrint("Render Complete")
	}

	return
}

// PostRender is called on all nodes once Render has returned.
func (rc *RenderContext) PostRender() error {
	// post process image
	for _, node := range rc.nodes {
		if err := node.PostRender(rc); err != nil {
			return err
		}
	}

	return nil
}

// GetMaterialId returns the shader id for the given node name.
func GetMaterialId(name string) int32 {
	return grc.GetMaterialId(name)
}

// GetMaterialId returns the shader id for the given node name.
//
// Deprecated: Should use the core API global GetMaterialId.
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

// AddNode adds a node to the core.
func (rc *RenderContext) AddNode(node Node) {
	rc.nodes = append(rc.nodes, node)
	rc.nodeMap[node.Name()] = node

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

// FindNode finds the node with the given name.
func (rc *RenderContext) FindNode(name string) Node {
	node, present := rc.nodeMap[name]

	if present {
		return node
	}

	return nil

}

// Error is called from the parser.
func (rc *RenderContext) Error(err error) error {
	log.Printf("Parse error: %v", err)
	return nil
}

// Node represents a node in the core system.
type Node interface {

	// Name returns the name of the node.
	Name() string

	// PreRender is called after loading scene and before render starts.  Nodes should
	// perform all init and may add other nodes in PreRender.
	PreRender(*RenderContext) error

	// PostRender is called after render is complete.
	PostRender(*RenderContext) error
}
