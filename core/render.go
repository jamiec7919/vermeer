// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/math/bluenoisedither"
	"github.com/jamiec7919/vermeer/math/ldseq"
	"log"
	"math"
	"math/rand"
	"sync"
)

type workitem struct {
	x, y, w, h int
}

//// Bit of a hack
type Image struct {
	PixelDelta [2]float32 // Size of pixel
}

// Framebuffer represents a buffer of pixels, RGB or deep.
//type Framebuffer struct {
//	Width, Height int
//	Buf           []float32
//}

// Aspect returns the aspect ratio of this framebuffer.
//func (fb *Framebuffer) Aspect() float32 {
//	return float32(fb.Width) / float32(fb.Height)
//}

var image *Image
var framebuffer *Framebuffer

// FrameAspect returns the aspect ratio of the current framebuffer.
func FrameAspect() float32 {
	return framebuffer.Aspect()
}

// FrameMetrics returns the width and height of the current framebuffer.
func FrameMetrics() (int, int) {
	return framebuffer.Width(), framebuffer.Height()
}

// FrameBuf returns the []float32 slice of pixels
func FrameBuf() []float32 {
	return nil // framebuffer.Buf
}

// render represents one goroutine.
func render(iter int, camera Camera, framebuffer *Framebuffer, work chan workitem, wg *sync.WaitGroup) {
	defer wg.Done()

	task := new(RenderTask)
	task.rand = rand.New(rand.NewSource(rand.Int63()))
	ray := task.NewRay()
	sc := task.NewShaderContext()
	sc.Image = image

	tile := make([]TraceSample, 32*32)

	for item := range work {
		for j := 0; j < item.h; j++ {
			for i := 0; i < item.w; i++ {

				x := i + item.x
				y := j + item.y

				//pixIdx := x + y*framebuffer.Width

				// If buffer size isn't a multiple of tile size skip any overhanging pixels
				if x >= framebuffer.Width() || y >= framebuffer.Height() {
					continue
				}

				_, rasterX, rasterY := ldseq.RasterXY(12, uint32(iter), uint32(x), uint32(y), 0, 0)
				//rasterX = rand.Float64() + float64(x)
				//rasterY = rand.Float64() + float64(y)

				ditheridx := (x % bluenoisedither.TileSize) + ((y % bluenoisedither.TileSize) * bluenoisedither.TileSize)

				time := ldseq.VanDerCorput(uint64(iter), bluenoisedither.Tile1D[0][ditheridx])
				lambda := (colour.LambdaMax-colour.LambdaMin)*ldseq.VanDerCorput(uint64(iter), bluenoisedither.Tile1D[1][ditheridx]) + colour.LambdaMin

				lensU := ldseq.VanDerCorput(uint64(iter), bluenoisedither.Tile2D[0][ditheridx][0])
				lensV := ldseq.Sobol(uint64(iter), bluenoisedither.Tile2D[0][ditheridx][1])

				if filter != nil {
					pixu := rasterX - math.Floor(rasterX)
					pixv := rasterY - math.Floor(rasterY)

					u, v := filter.WarpSample(pixu, pixv)

					rasterX = math.Floor(rasterX) + 0.5 + u
					rasterY = math.Floor(rasterY) + 0.5 + v
				}

				sc.X = int32(rasterX)
				sc.Y = int32(rasterY)

				w, h := FrameMetrics()

				sc.Sx = float32(-1.0 + 2.0*(rasterX/float64(w))) // note x [-filter.width/2,w+filter.width/2)
				sc.Sy = -float32(-1.0 + 2.0*(rasterY/float64(h)))

				sc.Lambda = float32(lambda)
				sc.Time = float32(time)

				camera.ComputeRay(sc, lensU, lensV, ray)

				ray.I = int(iter)
				ray.Scramble = bluenoisedither.Tile2D[1][ditheridx] //framescramble[pixIdx].scramble

				Trace(ray, &tile[i+j*32])

			}
		}

		framebuffer.AddSampleTile(item.x, item.y, iter, item.w, item.h, tile)
	}

	task.ReleaseRay(ray)
	task.ReleaseShaderContext(sc)
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

	image = &Image{}

	stats.begin()

	finish := false

	var iter int

	for !finish {

		if globals.MaxIter > 0 && iter >= globals.MaxIter-1 {
			finish = true
		}

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

		// TODO: Should check if stdout is to terminal or not.
		fmt.Printf("\rIter %v", iter)

		select {
		case <-exit:
			finish = true
		default:
		}

		iter++
	}

	framebuffer.Flush(iter)

	// Skip to next line after iter print.
	fmt.Printf("\n")

	stats.end()

	return stats, nil

}
