// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/jamiec7919/vermeer/math/ldseq"
	"log"
	"math"
	"math/rand"
	"sync"
)

// Need to pick scramble values per pixel (and per frame for raster positions). Basic
// values for lens (u&v), time, lambda.
// 32 byte.  Not sure I like this, a large buffer for such things!
type pixelscramble struct {
	lensU, lensV uint64
	time         uint64
	lambda       uint64
}

var framescramble []pixelscramble

type workitem struct {
	x, y, w, h int
}

//// Bit of a hack

// Framebuffer represents a buffer of pixels, RGB or deep.
type Framebuffer struct {
	Width, Height int
	Buf           []float32
}

// Aspect returns the aspect ratio of this framebuffer.
func (fb *Framebuffer) Aspect() float32 {
	return float32(fb.Width) / float32(fb.Height)
}

var framebuffer *Framebuffer

// FrameAspect returns the aspect ratio of the current framebuffer.
func FrameAspect() float32 {
	return framebuffer.Aspect()
}

// FrameMetrics returns the width and height of the current framebuffer.
func FrameMetrics() (int, int) {
	return framebuffer.Width, framebuffer.Height
}

// FrameBuf returns the []float32 slice of pixels
func FrameBuf() []float32 {
	return framebuffer.Buf
}

// render represents one goroutine.
func render(iter int, camera Camera, framebuffer *Framebuffer, work chan workitem, wg *sync.WaitGroup) {
	defer wg.Done()

	task := new(RenderTask)
	task.rand = rand.New(rand.NewSource(rand.Int63()))
	ray := task.NewRay()

	for item := range work {
		for j := 0; j < item.h; j++ {
			for i := 0; i < item.w; i++ {

				x := i + item.x
				y := j + item.y

				pixIdx := x + y*framebuffer.Width

				// If buffer size isn't a multiple of tile size skip any overhanging pixels
				if x >= framebuffer.Width || y >= framebuffer.Height {
					continue
				}

				_, rasterX, rasterY := ldseq.RasterXY(12, uint32(iter), uint32(x), uint32(y), 0, 0)
				//rasterX = rand.Float64() + float64(x)
				//rasterY = rand.Float64() + float64(y)

				time := ldseq.VanDerCorput(uint64(iter), framescramble[pixIdx].time)
				lambda := (720-450)*ldseq.VanDerCorput(uint64(iter), framescramble[pixIdx].lambda) + 450

				lensU := ldseq.VanDerCorput(uint64(iter), framescramble[pixIdx].lensU)
				lensV := ldseq.Sobol(uint64(iter), framescramble[pixIdx].lensV)

				if filter != nil {
					pixu := rasterX - math.Floor(rasterX)
					pixv := rasterY - math.Floor(rasterY)

					u, v := filter.WarpSample(pixu, pixv)

					rasterX = math.Floor(rasterX) + 0.5 + u
					rasterY = math.Floor(rasterY) + 0.5 + v
				}

				camera.ComputeRay(rasterX, rasterY, lensU, lensV, time, lambda, ray)

				samp := TraceSample{}
				ray.I = int(iter)
				Trace(ray, &samp)

				for k := 0; k < 3; k++ {
					framebuffer.Buf[(x+y*framebuffer.Width)*3+k] = (framebuffer.Buf[(x+y*framebuffer.Width)*3+k]*float32(iter) + samp.Colour[k]) / float32(iter+1)
				}

			}
		}
	}

	task.ReleaseRay(ray)
}

// Render is called to start the render process.
func Render(maxIter int, exit chan bool) (RenderStats, error) {
	log.Print("Begin Render")
	// Override globals with command line
	if maxIter > -1 {
		globals.MaxIter = maxIter
	}

	// 1. Find camera
	camName := "camera"

	if globals.Camera != "" {
		camName = globals.Camera
	}

	camNode := FindNode(camName)

	if camNode == nil {
		return stats, ErrNoCamera
	}

	camera, ok := camNode.(Camera)

	if !ok {
		return stats, ErrNoCamera
	}

	framescramble = make([]pixelscramble, framebuffer.Width*framebuffer.Height)

	for i := range framescramble {
		framescramble[i].lensU = uint64(rand.Int63())
		framescramble[i].lensV = uint64(rand.Int63())
		framescramble[i].time = uint64(rand.Int63())
		framescramble[i].lambda = uint64(rand.Int63())

	}

	stats.begin()

	finish := false

	for iter := 0; (iter < globals.MaxIter || globals.MaxIter == 0) && !finish; iter++ {

		// Spawn one goroutine per CPU (ish)
		workqueue := make(chan workitem)
		var wg sync.WaitGroup

		for i := 0; i < globals.MaxGoRoutines && i < 10; i++ {
			wg.Add(1)
			go render(iter+1, camera, framebuffer, workqueue, &wg)
		}

		// Parcel out frame tiles to the work queues.
		for j := 0; j < globals.YRes; j += 32 {
			for i := 0; i < globals.XRes; i += 32 {
				workqueue <- workitem{i, j, 32, 32}
			}
		}

		close(workqueue)
		wg.Wait()

		log.Printf("Iter %v", iter)

		select {
		case <-exit:
			finish = true
		default:
		}
	}

	stats.end()

	return stats, nil

}
