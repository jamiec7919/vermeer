// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filter

import (
	"github.com/jamiec7919/vermeer/core"
	"math"
)

// GaussianFilter implements the Gaussian pixel filter.
type GaussianFilter struct {
	NodeDef  core.NodeDef
	NodeName string `node:"Name"`
	Width    float32
	Res      int

	sampler *Sampler
}

// WarpSample implements core.PixelFilter.
func (f *GaussianFilter) WarpSample(r0, r1 float64) (u float64, v float64) {
	return f.sampler.WarpSample(r0, r1)
}

// Name is a core.Node method.
func (f *GaussianFilter) Name() string { return f.NodeName }

// Def is a core.Node method.
func (f *GaussianFilter) Def() core.NodeDef { return f.NodeDef }

// PreRender is a core.Node method.
func (f *GaussianFilter) PreRender() error {

	gauss := func(x, y float64) float64 {
		q := math.Sqrt(x*x + y*y)
		// Cutoff the filter
		if q > float64(f.Width/2) {
			return 0
		}

		sigma := float64(1.0 / math.Sqrt(float64(f.Width)))

		G := (1 / (2 * math.Pi * sqr(sigma))) * math.Exp(float64(-(x*x+y*y)/2*sqr(sigma)))

		return G
	}

	f.sampler = CreateSampler(f.Res, float64(f.Width), gauss)
	/*
		res := 61
		buf := make([]float32, res*res)
		u := -f.Width / 2
		v := -f.Width / 2
		du := f.Width / float32(res)
		dv := f.Width / float32(res)

		for j := 0; j < res; j++ {
			for i := 0; i < res; i++ {
				buf[i+j*res] = float32(gauss(float64(u), float64(v)))
				u += du
			}
			v += dv
			u = -f.Width / 2
		}
		fp, err := os.Create("outgauss.float")
		defer fp.Close()
		if err != nil {
			return err
		}
		err = binary.Write(fp, binary.LittleEndian, buf)
		if err != nil {
			return err
		}
	*/
	return nil
}

// PostRender is a core.Node method.
func (f *GaussianFilter) PostRender() error { return nil }
