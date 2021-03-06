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
	//"fmt"
	"bytes"
	"github.com/blezek/tga"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	_ "golang.org/x/image/tiff" // Imported for effect
	"image"
	_ "image/jpeg" // Imported for effect
	_ "image/png"  // Imported for effect
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// Uses https://github.com/jteeuwen/go-bindata
//go:generate go-bindata -nocompress -pkg=texture data/
const embeddedTexture = "data/checker.png"

// Texture represents a texture image.  (shouldn't be public)
type Texture struct {
	url    string
	fmt    int
	w, h   int
	data   []byte
	mipmap mipmap
}

// TexStore is the type of the cache.
type TexStore map[string]*Texture

var texStore atomic.Value

var loadMutex sync.Mutex

var testTexture *Texture

func init() {
	data, err := Asset(embeddedTexture)

	if err != nil {
		log.Printf("Failed to load embedded fallback texture: %v", err)

		tmp := CreateRGBTexture(2, 2)
		tmp.url = "test"
		tmp.SetRGB(0, 0, 2, 2, 2)
		tmp.SetRGB(1, 0, 250, 150, 250)
		tmp.SetRGB(0, 1, 250, 150, 250)
		tmp.SetRGB(1, 1, 2, 2, 2)

		mipmap := stdfilter(tmp.w, tmp.h, tmp.data, 3)

		tmp.mipmap = mipmap
		testTexture = tmp

	} else {
		t, err := loadTexture(embeddedTexture, bytes.NewReader(data))
		testTexture = t

		if err != nil {
			log.Printf("Failed to load embedded fallback texture: %v", err)
		}
	}
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

	return loadTexture(url, file)
}

func loadTexture(url string, file io.Reader) (*Texture, error) {

	var err error
	var m image.Image

	if filepath.Ext(url) == ".tga" || filepath.Ext(url) == ".TGA" {
		// Decode the image.
		m, err = tga.Decode(file)

	} else {
		// Decode the image.
		m, _, err = image.Decode(file)
	}

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

	mipmap := stdfilter(t.w, t.h, t.data, 3)

	t.mipmap = mipmap

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
		//loadMutex.Unlock()
		log.Printf("texture.SampleRGB: \"%v\": %v", filename, err)
		//return nil, err
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
func SampleRGB(filename string, sg *core.ShaderContext) (out [3]float32) {

	// This uses an atomic copy-on-write for the textures store
	//ds = m.Max(1, 1/ds)
	//dt = m.Max(1, 1/dt)

	//fmt.Printf("%v %v\n", ds, dt)
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

	deltaTx := m.Vec2Scale(sg.Image.PixelDelta[0], sg.Dduvdx)
	deltaTy := m.Vec2Scale(sg.Image.PixelDelta[1], sg.Dduvdy)

	deltaTx[0] = deltaTx[0] * float32(img.w)
	deltaTy[0] = deltaTy[0] * float32(img.w)

	deltaTx[1] = deltaTx[1] * float32(img.h)
	deltaTy[1] = deltaTy[1] * float32(img.h)

	ds := m.Max(m.Abs(deltaTx[0]), m.Abs(deltaTy[0]))
	dt := m.Max(m.Abs(deltaTx[1]), m.Abs(deltaTy[1]))

	ds = m.Vec2Length(deltaTx)
	dt = m.Vec2Length(deltaTy)
	//	loadMutex.Unlock()
	/*
		deltaTx := m.Vec2Scale(sg.Image.PixelDelta[0], sg.Dduvdx)
		deltaTy := m.Vec2Scale(sg.Image.PixelDelta[1], sg.Dduvdy)
	*/
	lod := m.Log2(m.Max(ds, dt))
	/*
		l := int(lod)

		if l > len(img.mipmap.mipmap)-1 {
			l = len(img.mipmap.mipmap) - 1
		}

		if l < 0 {
			l = 0
		}

		w := img.mipmap.mipmap[l].w
		h := img.mipmap.mipmap[l].h
		x := int(s * float32(w))
		y := int(t * float32(h))

		//fmt.Printf("%v %v %v %v\n", x, y, img.mipmap.mipmap[l].w, img.mipmap.mipmap[l].h)

		x = x % w
		y = y % h

		//fmt.Printf("%v %v\n", x, y)
		if x < 0 {
			x += w
		}
		if y < 0 {
			y += h
		}

		out[0] = float32(img.mipmap.mipmap[l].mipmap[(x+(y*img.mipmap.mipmap[l].w))*3+0]) / 255.0
		out[1] = float32(img.mipmap.mipmap[l].mipmap[(x+(y*img.mipmap.mipmap[l].w))*3+1]) / 255.0
		out[2] = float32(img.mipmap.mipmap[l].mipmap[(x+(y*img.mipmap.mipmap[l].w))*3+2]) / 255.0
	*/
	lod = lod

	if lod > float32(img.mipmap.MaxLevelOfDetail()) {
		lod = float32(img.mipmap.MaxLevelOfDetail())
	}

	if lod < 0 {
		lod = 0
	}

	out = img.mipmap.TrilinearSample(sg.U, sg.V, lod)
	out[0] /= 255.0
	out[1] /= 255.0
	out[2] /= 255.0

	//out[0] = lod * 10
	return
}
