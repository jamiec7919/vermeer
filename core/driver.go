// Copyright 2017 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

/*
Output of Vermeer is based on buffers of Arbitrary Output Variables.  A set of standard AOV buffers are defined and
shaders may also output other variables as required.
*/
const (
	AOVFloat32 int = iota
	AOVRGB
	AOVVec3
	AOVVec2
)

// Driver should attempt to write the given AOVs from the framebuffer in the appropriate format.
// Nodes should implement Driver and Write will be called during the PostRender phase of the rendering, after all
// other PostRender methods are called.
type Driver interface {
	Write(fb *Framebuffer, aovs []string) error
}

// Framebuffer represents a set of AOV buffers, all with the same width and height.
type Framebuffer struct {
	width, height int

	aovs map[string]interface{}
}

// AOV represents a named AOV buffer.
type AOV struct {
	Name string
	Type int
}

// NewFramebuffer creates a framebuffer with the given set of AOVs with w by h pixels.
func NewFramebuffer(w, h int, aovs []AOV) *Framebuffer {

	aovMap := map[string]interface{}{}

	for i := range aovs {
		switch aovs[i].Type {
		case AOVFloat32:
			aovMap[aovs[i].Name] = make([]float32, w*h)
		case AOVRGB:
			aovMap[aovs[i].Name] = make([]float32, w*h*3)
		}
	}

	return &Framebuffer{
		width:  w,
		height: h,
		aovs:   aovMap,
	}
}

// AOV returns the named AOV buffer (if it exists).
func (fb *Framebuffer) AOV(aov string) interface{} {
	return fb.aovs[aov]
}

// Width returns the width of the framebuffer in pixels.
func (fb *Framebuffer) Width() int {
	return fb.width
}

// Height returns the height of the framebuffer in pixels.
func (fb *Framebuffer) Height() int {
	return fb.height
}

// Aspect returns the aspect ratio width/height.
func (fb *Framebuffer) Aspect() float32 {
	return float32(fb.Width()) / float32(fb.Height())
}

// Flush is called with the last iteration count after all rendering is complete.
func (fb *Framebuffer) Flush(i int) {
}

// AddSample adds the i'th sample to the buffer at pixel (x,y).
// This updates all relevant AOV buffers at the same time.
func (fb *Framebuffer) AddSample(x, y, i int, s TraceSample) {
	rgb := fb.aovs["RGB"]

	if rgb != nil {
		buf := rgb.([]float32)

		buf[(x+(y*fb.width))*3+0] += (s.Colour[0] - buf[(x+(y*fb.width))*3+0]) / float32(i)
		buf[(x+(y*fb.width))*3+1] += (s.Colour[1] - buf[(x+(y*fb.width))*3+1]) / float32(i)
		buf[(x+(y*fb.width))*3+2] += (s.Colour[2] - buf[(x+(y*fb.width))*3+2]) / float32(i)
	}
}

// AddSample adds a tile of the i'th samples to the buffer starting at pixel (x,y).  The tile has size w x h.
// This updates all relevant AOV buffers at the same time.
func (fb *Framebuffer) AddSampleTile(x, y, i, w, h int, s []TraceSample) {
	rgb := fb.aovs["RGB"]

	if rgb != nil {
		buf := rgb.([]float32)

		for row := 0; row < h; row++ {
			for j := 0; j < w; j++ {
				buf[(x+j+((y+row)*fb.width))*3+0] += (s[j+(row*w)].Colour[0] - buf[(x+j+((y+row)*fb.width))*3+0]) / float32(i)
				buf[(x+j+((y+row)*fb.width))*3+1] += (s[j+(row*w)].Colour[1] - buf[(x+j+((y+row)*fb.width))*3+1]) / float32(i)
				buf[(x+j+((y+row)*fb.width))*3+2] += (s[j+(row*w)].Colour[2] - buf[(x+j+((y+row)*fb.width))*3+2]) / float32(i)
			}
		}
	}
}
