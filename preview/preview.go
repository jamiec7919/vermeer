// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package preview

import (
	"github.com/go-gl/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/jamiec7919/vermeer/internal/core"
	"log"
	"runtime"
)

var w, h int
var textures []gl.Texture
var window *glfw.Window

func initGL() error {
	gl.ClearColor(0, 0, 0, 0)
	gl.ClearDepth(1)
	gl.Enable(gl.TEXTURE_2D)
	gl.Disable(gl.DEPTH_TEST)

	textures = make([]gl.Texture, 1)

	gl.GenTextures(textures)

	if gl.GetError() != gl.NO_ERROR {
		log.Printf("TEX: Error glGenTextures %v %v", len(textures), gl.GetError())
	}

	// Texture 1
	textures[0].Bind(gl.TEXTURE_2D)

	buf := make([]uint8, 3*256*256)
	for i := range buf {
		buf[i] = uint8(i % 255)
	}
	updateTexture(256, 256, buf)
	return nil
}

func updateTexture(w, h int, buf []uint8) {
	textures[0].Bind(gl.TEXTURE_2D)

	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, w, h, 0, gl.RGB, gl.UNSIGNED_BYTE, buf)

	if gl.GetError() != gl.NO_ERROR {
		log.Printf("TEX: Error glTexImage2D %v %v %x", w, h, gl.GetError())
	}

}

func Run(rc *core.RenderContext) {
	defer glfw.Terminate()

	rc.PreviewChan = make(chan core.PreviewFrame)

	running := true

	initGL()

	for running && !window.ShouldClose() {
		select {
		case frame := <-rc.PreviewChan:
			updateTexture(frame.W, frame.H, frame.Buf)
		default:
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
			gl.Enable(gl.TEXTURE_2D)
			textures[0].Bind(gl.TEXTURE_2D)

			gl.Begin(gl.QUADS)
			gl.TexCoord2f(0, 0)
			gl.Vertex3f(-1, -1, 0)
			gl.TexCoord2f(0, 1)
			gl.Vertex3f(-1, 1, 0)
			gl.TexCoord2f(1, 1)
			gl.Vertex3f(1, 1, 0)
			gl.TexCoord2f(1, 0)
			gl.Vertex3f(1, -1, 0)
			gl.End()

			//running = glfw.Key(glfw.KeyEscape) == 0
			window.SwapBuffers()
			glfw.PollEvents()
		}
	}
}

func Init() (err error) {

	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		return err
	}

	window, err = glfw.CreateWindow(256, 256, "Vermeer Light Tools", nil, nil)

	if err != nil {
		return err
	}

	window.MakeContextCurrent()

	window.SetSizeCallback(onResize)

	return nil
}

func onResize(window *glfw.Window, iw, ih int) {
	w = iw
	h = ih
	//log.Printf("resized: %dx%d\n", w, h)
	gl.Viewport(0, 0, w, h)
}
