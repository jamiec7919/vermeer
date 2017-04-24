// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	//"fmt"
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/core/param"
	"github.com/jamiec7919/vermeer/texture"
	"net/url"
	"strconv"
)

type Texture struct {
	Filename string
	Chan     int
}

func (c *Texture) Float32(sg *core.ShaderContext) float32 {
	return texture.SampleFeline(c.Filename, sg)[c.Chan]
}

func (c *Texture) RGB(sg *core.ShaderContext) colour.RGB {

	return colour.RGB(texture.SampleFeline(c.Filename, sg))

}

type TextureTrilinear struct {
	Filename string
	Chan     int
}

func (c *TextureTrilinear) Float32(sg *core.ShaderContext) float32 {
	return texture.SampleRGB(c.Filename, sg)[c.Chan]
}

func (c *TextureTrilinear) RGB(sg *core.ShaderContext) colour.RGB {
	return colour.RGB(texture.SampleRGB(c.Filename, sg))

}

func CreateFloat32TextureMap(filename string) (param.Float32Uniform, error) {
	u, err := url.Parse(filename)

	if err != nil {
		return nil, err
	}

	chString := u.Query().Get("ch")

	ch, _ := strconv.Atoi(chString)

	filter := u.Query().Get("filter")

	switch filter {
	case "trilinear":
		return &TextureTrilinear{Filename: u.Path, Chan: ch}, nil
	}

	return &Texture{Filename: u.Path, Chan: ch}, nil

}

func CreateRGBTextureMap(filename string) (param.RGBUniform, error) {
	u, err := url.Parse(filename)

	if err != nil {
		return nil, err
	}

	filter := u.Query().Get("filter")

	switch filter {
	case "trilinear":
		return &TextureTrilinear{Filename: u.Path, Chan: 0}, nil
	}

	return &Texture{Filename: u.Path, Chan: 0}, nil
}
