// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	Package hdr implements implements the Radiance HDR format, supports only float32 RGB pixels.
*/
package hdr

import (
	"bufio"
	"math"
	"os"
	"vermeer/image"
)

const colourExcess = 128
const minEncodingLen = 8
const maxEncodingLen = 0x7fff
const minRunLen = 4

/*
type HDR struct {
	w, h   int
	buffer []byte
}
*/
func Open(filename string) (*HDRReader, error) {

	h := HDRReader{}

	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	h.file = file
	h.reader = bufio.NewReader(h.file)

	if err := h.readHeaders(); err != nil {
		return nil, err
	}

	return &h, nil
}

func convertComponent(expo uint, val byte) float32 {
	v := float32(val) / 256.0
	d := 1 << expo // float32(math.Pow(2.0, float64(expo)))
	return v * float32(d)
}

/*
func (i *HDR) At(x, y int) (r, g, b float32) {
	if x < 0 || x > i.w-1 || y < 0 || y > i.h-1 {
		return
	}

	or := i.buffer[(y*i.w*4)+(x*4)+0]
	og := i.buffer[(y*i.w*4)+(x*4)+1]
	ob := i.buffer[(y*i.w*4)+(x*4)+2]
	oe := i.buffer[(y*i.w*4)+(x*4)+3]

	expo := int(oe) - 128

	r = convertComponent(expo, int(or))
	g = convertComponent(expo, int(og))
	b = convertComponent(expo, int(ob))
	return
}
*/
func RGBEOfRGB(r, g, b float32) (or, og, ob, oe byte) {

	d := r

	if g > d {
		d = g
	}
	if b > d {
		d = b
	}

	if d < 0.000001 {
		return 0, 0, 0, 0
	}
	nd, e := math.Frexp(float64(d))
	n := float32(nd)

	df := n * 255.999 / d

	or = byte(r * df)
	og = byte(g * df)
	ob = byte(b * df)
	oe = byte(e + colourExcess)
	return

}

/*
func ReadHDR(filename string) (img *HDR, err error) {
	fin, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fin.Close()

	// Read header
	reader := bufio.NewReader(fin)
	// bytes := make([]byte, DefaultBufferSize)

	for {

		line, err := reader.ReadString('\n')

		if err != nil {
			return nil, err
		}
		if len(line) == 1 { // includes delim
			break
		}

		// log.Printf("%v", line)
	}

	line, err := reader.ReadString('\n')

	if err != nil {
		return nil, err
	}

	var xs, ys string
	var w, h int
	fmt.Sscanf(line, "%s %d %s %d", &ys, &h, &xs, &w)

	// log.Printf("%v %v %v %v", ys, h, xs, w)

	img = &HDR{
		w:      w,
		h:      h,
		buffer: make([]byte, w*h*4),
	}

	for j := 0; j < h; j++ {
		if err := readScanline(reader, img.buffer[j*w*4:j*w*4+w*4]); err != nil {
			return nil, err
		}
	}

	return img, nil
}
*/

func init() {

	open := func(filename string) (image.Reader, error) {
		r, err := Open(filename)

		if err == nil && r != nil {
			return r, err
		}
		return nil, err
	}

	image.RegisterReader(open)
}
