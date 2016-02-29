// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/cheggaaa/pb"
	"github.com/jamiec7919/vermeer/internal/nodeparser"
	"log"
	"math/rand"
	"sync"
	"time"
)

const TILESIZE = 64
const MAXGOROUTINES = 5
const NSAMP = 64

type Frame struct {
	w, h   int
	du, dv float32
	camera Camera
	scene  *Scene
	bar    *pb.ProgressBar
}

type Scene struct {
	prims  []Primitive
	lights []Light
}

type RenderContext struct {
	globals Globals
	imgbuf  []float32
	frames  []Frame
	nodes   []Node
	scene   Scene
	cameras []Camera
}

func NewRenderContext() *RenderContext {
	rc := &RenderContext{}
	rc.globals.XRes = 256
	rc.globals.YRes = 256
	rc.globals.UseProgress = true
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

	rc.imgbuf = make([]float32, 3*rc.globals.XRes*rc.globals.YRes)
	return nil
}

type WorkItem struct {
	x, y, w, h int
	samples    []float32
}

type RayData struct{}

// NOTE: we return the raydata here even though it is ignored in order to ensure that ray is
// heap allocated (for alignment purposes)
func renderFunc(frame *Frame, c chan *WorkItem, done chan *WorkItem, wg *sync.WaitGroup) *RayData {
	defer wg.Done()
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	ray := &RayData{}
	for w := range c {
		for j := 0; j < w.h; j++ {
			for i := 0; i < w.w; i++ {
				for k := 0; k < NSAMP; k++ {
					r0 := rnd.Float32()
					r1 := rnd.Float32()

					u := (float32(i+w.x) + r0) * frame.du
					v := (float32(j+w.y) + r1) * frame.dv

					P, D := frame.camera.ComputeRay(-1+u, 1-v, rnd)

					w.samples[(i+(j*w.w))*3] += P[0]
					w.samples[(i+(j*w.w))*3+1] += D[0]
					w.samples[(i+(j*w.w))*3+2] += 0

				}
				if frame.bar != nil {
					frame.bar.Increment()
				}
			}
		}

		for i := range w.samples {
			w.samples[i] /= NSAMP
		}

		done <- w
	}

	return ray
}

func (rc *RenderContext) Render() error {
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

	frame.scene = &rc.scene
	frame.w = rc.globals.XRes
	frame.h = rc.globals.YRes
	frame.du = 2.0 / float32(frame.w)
	frame.dv = 2.0 / float32(frame.h)

	if rc.globals.UseProgress {
		frame.bar = pb.StartNew(rc.globals.XRes * rc.globals.YRes)
	}

	var wg sync.WaitGroup
	workChan := make(chan *WorkItem)
	done := make(chan *WorkItem)

	for n := 0; n < rc.globals.MaxGoRoutines; n++ {
		wg.Add(1)
		go renderFunc(&frame, workChan, done, &wg)
	}

	complete := make(chan []float32)
	go func() {
		buf := make([]float32, frame.w*frame.h*3)
		var q []*WorkItem
		for d := range done {
			q = append(q, d)
		}

		for k := range q {
			for j := 0; j < q[k].h; j++ {
				for i := 0; i < q[k].w; i++ {
					rc.imgbuf[((i+q[k].x)+(j+q[k].y)*frame.w)*3+0] = q[k].samples[(i+(j*q[k].w))*3+0]
					rc.imgbuf[((i+q[k].x)+(j+q[k].y)*frame.w)*3+1] = q[k].samples[(i+(j*q[k].w))*3+1]
					rc.imgbuf[((i+q[k].x)+(j+q[k].y)*frame.w)*3+2] = q[k].samples[(i+(j*q[k].w))*3+2]
				}
			}
		}

		complete <- buf
	}()

	for j := 0; j < frame.h; j += TILESIZE {
		for i := 0; i < frame.w; i += TILESIZE {

			workChan <- &WorkItem{x: i, y: j, w: TILESIZE, h: TILESIZE, samples: make([]float32, TILESIZE*TILESIZE*3)}
		}
	}

	close(workChan)
	wg.Wait()
	close(done)

	rc.imgbuf = <-complete

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

func (rc *RenderContext) AddNode(node Node) {
	rc.nodes = append(rc.nodes, node)

	switch t := node.(type) {
	case Camera:
		rc.cameras = append(rc.cameras, t)
	case Primitive:
		rc.scene.prims = append(rc.scene.prims, t)
	case Light:
		rc.scene.lights = append(rc.scene.lights, t)

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

func (rc *RenderContext) Dispatch(method string, _params map[string]interface{}) error {
	params := Params(_params)

	create, present := nodeTypes[method]

	if present {
		return create(rc, params)
	}

	return ErrNodeNotRegistered
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

var nodeTypes = map[string]func(*RenderContext, Params) error{}

func RegisterNodeType(name string, create func(*RenderContext, Params) error) error {
	if nodeTypes[name] == nil {
		nodeTypes[name] = create
		return nil
	}
	return ErrNodeAlreadyRegistered
}
