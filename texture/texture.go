// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package texture implements an efficient texture cache.

At present it does virtually nothing other than thread-safe loading of texture files.  Doesn't
release memory or attempt to use tiled caching.

Textures are currently represented simply by a string which references into a hashmap.  Lookup sounds
inefficient but has never shown up as significant on profiling.  Expected to change as many more
textures are used in shaders. */
package texture

import (
	//_ "github.com/ftrvxmtrx/tga"
	_ "golang.org/x/image/tiff" // Imported for effect
	"image"
	_ "image/jpeg" // Imported for effect
	_ "image/png"  // Imported for effect
	"log"
	"os"
	"sync"
	"sync/atomic"
)

// Texture represents a texture image.  (shouldn't be public)
type Texture struct {
	url  string
	fmt  int
	w, h int
	data []byte
}

// TexStore is the type of the cache.
type TexStore map[string]*Texture

var texStore atomic.Value

var loadMutex sync.Mutex

var testTexture *Texture

func init() {
	tmp := CreateRGBTexture(2, 2)
	tmp.url = "test"
	tmp.SetRGB(0, 0, 250, 250, 250)
	tmp.SetRGB(1, 0, 250, 5, 250)
	tmp.SetRGB(0, 1, 250, 5, 250)
	tmp.SetRGB(1, 1, 250, 250, 250)
	testTexture = tmp
	texStore.Store(make(TexStore))
}

/*
Correct way to get an RGBA image:
b := src.Bounds()
m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
draw.Draw(m, m.Bounds(), src, b.Min, draw.Src)
*/

// LoadTexture returns a texture object or an error if can't be openend.  Takes a url for
// future network texture server. (shouldn't be public)
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
			//t.SetRGB(i, m.Bounds().Max.Y-1-j, uint8(r), uint8(g), uint8(b))
			//t.SetRGB(i, m.Bounds().Max.Y-1-j, uint8(r/alpha), uint8(g/alpha), uint8(b/alpha))
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

// SetRGB sets a pixel in a Texture object. (shouldn't be public)
func (tex *Texture) SetRGB(x, y int, r, g, b byte) {
	tex.data[(x+(y*tex.w))*3+0] = r
	tex.data[(x+(y*tex.w))*3+1] = g
	tex.data[(x+(y*tex.w))*3+2] = b
}

// CreateRGBTexture creates an RGB texture of appropriate size.
func CreateRGBTexture(w, h int) *Texture {
	return &Texture{
		w:    w,
		h:    h,
		data: make([]byte, w*h*3),
	}
}

func cacheMiss(filename string) (*Texture, error) {
	loadMutex.Lock()
	defer loadMutex.Unlock()

	// Load current version, make sure the previous locker hasn't loaded the
	// same image we want.
	textures := texStore.Load().(TexStore)
	if img, present := textures[filename]; present {
		return img, nil
	}

	tex, err := LoadTexture(filename)

	if err != nil {
		loadMutex.Unlock()
		log.Printf("texture.SampleRGB: \"%v\": %v", filename, err)
		return nil, err
	}

	texturesNew := make(TexStore)

	for k, v := range textures {
		texturesNew[k] = v
	}
	texturesNew[filename] = tex
	texStore.Store(texturesNew)

	return tex, nil

}

// SampleRGB samples an RGB value from the given file using the coords s,t and footprint ds,dt.
func SampleRGB(filename string, s, t, ds, dt float32) (out [3]float32) {
	// This uses an atomic copy-on-write for the textures store

	//loadMutex.Lock()
	textures := texStore.Load().(TexStore)
	img := textures[filename]

	if img == nil {
		img2, err := cacheMiss(filename)

		if err != nil {
			return
		}

		img = img2
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
