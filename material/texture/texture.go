// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package texture

import (
	"image"
	_ "image/jpeg"
	"log"
	"os"
	"sync"
)

type Texture struct {
	url  string
	fmt  int
	w, h int
	data []byte
}

var loadMutex sync.Mutex
var texMutex sync.RWMutex

var textures = map[string]*Texture{}

var testTexture *Texture

func init() {
	tmp := CreateRGBTexture(2, 2)
	tmp.url = "test"
	tmp.SetRGB(0, 0, 250, 250, 250)
	tmp.SetRGB(1, 0, 250, 5, 250)
	tmp.SetRGB(0, 1, 250, 5, 250)
	tmp.SetRGB(1, 1, 250, 250, 250)
	testTexture = tmp
}

/*
Correct way to get an RGBA image:
b := src.Bounds()
m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
draw.Draw(m, m.Bounds(), src, b.Min, draw.Src)
*/

func LoadTexture(url string) (*Texture, error) {

	file, err := os.Open(url)
	if err != nil {
		return testTexture, err
	}
	defer file.Close()

	// Decode the image.
	m, _, err := image.Decode(file)
	if err != nil {
		return testTexture, err
	}

	t := CreateRGBTexture(m.Bounds().Max.X-m.Bounds().Min.X, m.Bounds().Max.Y-m.Bounds().Min.Y)
	t.url = url

	//		tex := AllocTexImage(m.Bounds().Max.X-m.Bounds().Min.X, m.Bounds().Max.Y-m.Bounds().Min.Y)

	for j := m.Bounds().Min.Y; j < m.Bounds().Max.Y; j++ {
		for i := m.Bounds().Min.X; i < m.Bounds().Max.X; i++ {
			r, g, b, _ := m.At(i, j).RGBA()
			t.SetRGB(i, m.Bounds().Max.Y-1-j, uint8(r>>8), uint8(g>>8), uint8(b>>8))
			//				tex.Set(i, m.Bounds().Max.Y-1-j, byte(r), byte(g), byte(b))
		}
	}
	/*
		t.rset = make(map[int]*TexImage)
		t.rset[t.genHash(log2(tex.w), log2(tex.h))] = tex

		w := tex.w

		for w > 0 {
			h := tex.h

			texa := tex

			for h > 0 {
				tex2 := AllocTexImage(w, h)
				tex2.Resample(texa)
				t.rset[t.genHash(log2(w), log2(h))] = tex2
				texa = tex2
				/// log.Printf("Gen %v %v", log2(w), log2(h))
				h /= 2
			}
			w /= 2
		}
	*/ /*
		tex3 := AllocTexImage(tex2.w/2, tex2.h/2)
		tex3.Resample(tex2)
		return tex3.Texture(), nil
	*/
	return t, nil
}

func (tex *Texture) SetRGB(x, y int, r, g, b byte) {
	tex.data[(x+(y*tex.w))*3+0] = r
	tex.data[(x+(y*tex.w))*3+1] = g
	tex.data[(x+(y*tex.w))*3+2] = b
}

func CreateRGBTexture(w, h int) *Texture {
	return &Texture{
		w:    w,
		h:    h,
		data: make([]byte, w*h*3),
	}
}

func SampleRGB(filename string, s, t, ds, dt float32) (out [3]float32) {

	//loadMutex.Lock()
	texMutex.RLock()
	img := textures[filename]
	texMutex.RUnlock()

	if img == nil {
		loadMutex.Lock()
		tex, err := LoadTexture(filename)

		if err != nil {
			loadMutex.Unlock()
			log.Printf("texture.SampleRGB: %v", err)
			return
		}

		texMutex.Lock()
		textures[filename] = tex
		texMutex.Unlock()
		loadMutex.Unlock()
		img = tex
	}
	//	loadMutex.Unlock()

	x := int(s * float32(img.w))
	y := int(t * float32(img.h))

	x = x % img.w
	y = y % img.h

	if x < 0 {
		x += img.w
	}
	if y < 0 {
		y += img.h
	}

	out[0] = float32(img.data[(x+(y*img.w))*3+0]) / 255.0
	out[1] = float32(img.data[(x+(y*img.w))*3+1]) / 255.0
	out[2] = float32(img.data[(x+(y*img.w))*3+2]) / 255.0

	return
}
