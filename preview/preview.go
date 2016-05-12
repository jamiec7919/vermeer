// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package preview

import (
	"errors"
	"fmt"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/jamiec7919/vermeer/core"
	"runtime"
)

var w, h int
var textures []gl.Texture
var window *glfw.Window

type Preview struct {
	frameChan chan core.PreviewFrame
	window    *glfw.Window
}

func (p *Preview) UpdateFrame(frame core.PreviewFrame) {
	if p.window != nil {
		glfw.PostEmptyEvent()
		p.frameChan <- frame
	}
}

// Close is called by RenderContext when rendering should end
func (p *Preview) Close() {
	if p.window != nil {
		p.window.SetShouldClose(true)
		close(p.frameChan)
		glfw.PostEmptyEvent()
	}
}

func initGL() error {
	gl.ClearColor(0, 0, 0, 0)
	gl.ClearDepth(1)
	gl.Enable(gl.TEXTURE_2D)
	gl.Disable(gl.DEPTH_TEST)

	textures = make([]gl.Texture, 1)

	gl.GenTextures(textures)

	if gl.GetError() != gl.NO_ERROR {
		return errors.New(fmt.Sprintf("Error glGenTextures %v %v", len(textures), gl.GetError()))
	}

	// Texture 1
	textures[0].Bind(gl.TEXTURE_2D)

	buf := make([]uint8, 3*256*256)
	for i := range buf {
		buf[i] = uint8(i % 255)
	}

	return updateTexture(256, 256, buf)
}

func updateTexture(w, h int, buf []uint8) error {
	textures[0].Bind(gl.TEXTURE_2D)

	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, w, h, 0, gl.RGB, gl.UNSIGNED_BYTE, buf)

	if gl.GetError() != gl.NO_ERROR {
		return errors.New(fmt.Sprintf("Error glTexImage2D %v %v %v", w, h, gl.GetError()))
	}

	return nil
}

func redraw() {
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

}

// This needs to be run in the main thread (and locked to it)
func (p *Preview) Run() error {
	defer glfw.Terminate()

	running := true

	if err := initGL(); err != nil {
		return err
	}

	redraw()

	for running && !p.window.ShouldClose() {
		select {
		case frame := <-p.frameChan:
			updateTexture(frame.W, frame.H, frame.Buf)
			redraw()
		default:
		}
		glfw.WaitEvents()

	}
	p.window.Destroy()
	p.window = nil
	runtime.UnlockOSThread()
	return nil
}

func Init() (preview *Preview, err error) {

	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		return nil, err
	}

	window, err = glfw.CreateWindow(1024, 1024, "Vermeer Light Tools", nil, nil)

	if err != nil {
		return nil, err
	}

	window.MakeContextCurrent()

	window.SetSizeCallback(onResize)

	p := &Preview{make(chan core.PreviewFrame), window}

	return p, nil
}

func onResize(window *glfw.Window, iw, ih int) {
	w = iw
	h = ih
	//log.Printf("resized: %dx%d\n", w, h)
	gl.Viewport(0, 0, w, h)
}
