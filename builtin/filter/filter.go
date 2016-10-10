// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filter

import (
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/nodes"
)

func sqr(x float64) float64 { return x * x }

func init() {
	nodes.Register("GaussianFilter", func() (core.Node, error) {
		out := GaussianFilter{Res: 17, Width: 2}

		return &out, nil
	})
	nodes.Register("AiryFilter", func() (core.Node, error) {
		out := AiryFilter{Res: 49, Width: 6, Peak: 4}

		return &out, nil
	})
}

// Sampler implements 2D sampling using a filter function as PDF.
type Sampler struct {
	filter []float64
	pdf    [][]float64
	pV     []float64
	pVU    [][]float64
	cdfV   []float64
	cdfVU  [][]float64
	n      int
	w      float64
}

// WarpSample warps the given sample in [0,1]x[0,1] by the filter PDF.
func (fil *Sampler) WarpSample(r0, r1 float64) (u float64, v float64) {
	// search pV for r0:
	uI := -1

	for i := 0; i < len(fil.cdfV); i++ {
		uI = i
		if r0 < fil.cdfV[i] {

			// Lerp between stored CDF values to avoid discretisation errors.
			if i == 0 {
				du := r0 / fil.cdfV[i]
				u = float64((-fil.w / 2)) + (du / fil.w)
			} else {
				du := (r0 - fil.cdfV[i-1]) / (fil.cdfV[i] - fil.cdfV[i-1])
				u = float64((-fil.w / 2) + fil.w*(float64(i)+du)/float64(fil.n-1))

			}
			break
		}
	}

	for i := 0; i < len(fil.cdfVU[uI]); i++ {
		if r1 < fil.cdfVU[uI][i] {

			// Lerp between stored CDF values to avoid discretisation errors.
			if i == 0 {
				dv := r1 / fil.cdfVU[uI][i]
				v = float64((-fil.w / 2)) + (dv / fil.w)
			} else {
				dv := (r1 - fil.cdfVU[uI][i-1]) / (fil.cdfVU[uI][i] - fil.cdfVU[uI][i-1])
				v = float64((-fil.w / 2) + fil.w*(float64(i)+dv)/float64(fil.n-1))

			}

			// These snap to grid points
			//	u = float64((-fil.w / 2) + fil.w*float64(uI)/float64(fil.n-1))
			//v = float64((-fil.w / 2) + fil.w*float64(i)/float64(fil.n-1))
			return
		}
	}

	return 0, 0
}

// CreateSampler will return a sampled version of the filter function given by f.
// w is the width of the filter (i.e. x in [-w/2,w/2])
// n is the resolution of the sampling.
// The Sampler thus returned can be used to sample locations using the filter values as PDF.
func CreateSampler(n int, w float64, f func(x, y float64) float64) (fil *Sampler) {
	fil = &Sampler{}

	fil.n = n
	fil.w = w

	fil.filter = make([]float64, n*n)

	du := w / float64(n-1)
	u := -w / 2

	F := float64(0)

	for j := 0; j < n; j++ {
		dv := w / float64(n-1)
		v := -w / 2
		for i := 0; i < n; i++ {
			fuv := f(u, v)
			fil.filter[j+(i*n)] = fuv
			F += fuv

			v += dv
		}
		u += du
	}

	// log.Printf("%v", fil.filter)

	fil.pdf = make([][]float64, n)

	for j := 0; j < n; j++ {
		fil.pdf[j] = make([]float64, n)

		for i := 0; i < n; i++ {
			fil.pdf[j][i] = fil.filter[j+(i*n)] / F
		}
	}

	// log.Printf("%v", fil.pdf)

	fil.pV = make([]float64, n)

	for j := 0; j < n; j++ {

		for i := 0; i < n; i++ {
			fil.pV[j] += fil.pdf[j][i]
		}
	}

	// log.Printf("pV: %v", fil.pV)

	fil.pVU = make([][]float64, n)

	for j := 0; j < n; j++ {
		fil.pVU[j] = make([]float64, n)

		for i := 0; i < n; i++ {
			fil.pVU[j][i] = fil.pdf[j][i] / fil.pV[j]
		}
	}

	p := float64(0)
	fil.cdfV = make([]float64, n)
	for i := 0; i < n; i++ {
		p += fil.pV[i]
		fil.cdfV[i] = p
	}

	// log.Printf("cdfV: %v", fil.cdfV)

	fil.cdfVU = make([][]float64, n)
	for j := 0; j < n; j++ {
		fil.cdfVU[j] = make([]float64, n)

		p := float64(0)

		for i := 0; i < n; i++ {
			p += fil.pVU[j][i]
			fil.cdfVU[j][i] = p
		}
	}

	return
	// Finally, select u by cdfV, then select v from cdf_uv
	// using the appropriate index
}
