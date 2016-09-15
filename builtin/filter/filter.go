// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filter

import (
	//	m "github.com/jamiec7919/vermeer/math"
	//"encoding/binary"
	//"fmt"
	"math"
	//"os"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/nodes"
)

func sqr(x float64) float64 { return x * x }

type GaussianFilter struct {
	NodeDef  core.NodeDef
	NodeName string `node:"Name"`
	Width    float32
	Res      int

	sampler *FilterSampler
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

	f.sampler = CreateFilterSampler(f.Res, float64(f.Width), gauss)
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

type AiryFilter struct {
	NodeDef  core.NodeDef `node:"-"`
	NodeName string       `node:"Name"`
	Width    float32      `node:",opt"`
	Res      int          `node:",opt"`
	Peak     float32      `node:",opt"`

	sampler *FilterSampler
}

// WarpSample implements core.PixelFilter.
func (f *AiryFilter) WarpSample(r0, r1 float64) (u float64, v float64) {
	return f.sampler.WarpSample(r0, r1)
}

// Name is a core.Node method.
func (f *AiryFilter) Name() string { return f.NodeName }

// Def is a core.Node method.
func (f *AiryFilter) Def() core.NodeDef { return f.NodeDef }

//www.atnf.csiro.au/computing/software/gipsy/sub/bessel.c
/*------------------------------------------------------------*/
/* PURPOSE: Evaluate Bessel function of first kind and order  */
/*          1 at input x                                      */
/*------------------------------------------------------------*/
func BesselJ1(x float64) float64 {

	ax := math.Abs(x)

	if ax < 8.0 {
		y := x * x

		ans1 := x * (72362614232.0 + y*(-7895059235.0+y*(242396853.1+
			y*(-2972611.439+y*(15704.48260+y*(-30.16036606))))))

		ans2 := 144725228442.0 + y*(2300535178.0+y*(18583304.74+
			y*(99447.43394+y*(376.9991397+y*1.0))))

		return ans1 / ans2

	}

	z := 8.0 / ax
	y := z * z
	xx := ax - 2.356194491

	ans1 := 1.0 + y*(0.183105e-2+y*(-0.3516396496e-4+
		y*(0.2457520174e-5+y*(-0.240337019e-6))))

	ans2 := 0.04687499995 + y*(-0.2002690873e-3+
		y*(0.8449199096e-5+y*(-0.88228987e-6+
			y*0.105787412e-6)))

	ans := math.Sqrt(0.636619772/ax) * (math.Cos(xx)*ans1 - z*math.Sin(xx)*ans2)

	if x < 0.0 {
		ans = -ans
	}

	return ans

}

// PreRender is a core.Node method.
func (f *AiryFilter) PreRender() error {

	airy := func(x, y float64) float64 {
		// http://www.prasa.org/proceedings/2012/prasa2012-13.pdf
		q := math.Sqrt(x*x + y*y)

		// Cutoff the filter
		if q > float64(f.Width/2) {
			return 0
		}

		lambda := float64(550)
		N := 5.6 // F stop

		// 20000 is a kludge to make it fit into reasonable pixel area.  This should probably be calculated from
		// actual pixel sizes.
		v := (20000 / float64(f.Width)) * (math.Pi * q) / (lambda * N)

		A := float64(f.Peak) * sqr(2*BesselJ1(v)/v)

		//fmt.Printf("%v %v %v\n", A, v, BesselJ1(v))
		return A
	}

	f.sampler = CreateFilterSampler(f.Res, float64(f.Width), airy)

	/*
		res := 61
		buf := make([]float32, res*res)
		u := -f.Width / 2
		v := -f.Width / 2
		du := f.Width / float32(res)
		dv := f.Width / float32(res)

		for j := 0; j < res; j++ {
			for i := 0; i < res; i++ {
				buf[i+j*res] = float32(airy(float64(u), float64(v)))
				u += du
			}
			v += dv
			u = -f.Width / 2
		}
			fp, err := os.Create("out.float")
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
func (f *AiryFilter) PostRender() error { return nil }

type FilterSampler struct {
	filter []float64
	pdf    [][]float64
	p_v    []float64
	p_vu   [][]float64
	cdf_v  []float64
	cdf_vu [][]float64
	n      int
	w      float64
}

func (fil *FilterSampler) WarpSample(r0, r1 float64) (u float64, v float64) {
	// search p_v for r0:
	u_i := -1

	for i := 0; i < len(fil.cdf_v); i++ {
		u_i = i
		if r0 < fil.cdf_v[i] {

			// Lerp between stored CDF values to avoid discretisation errors.
			if i == 0 {
				du := r0 / fil.cdf_v[i]
				u = float64((-fil.w / 2)) + (du / fil.w)
			} else {
				du := (r0 - fil.cdf_v[i-1]) / (fil.cdf_v[i] - fil.cdf_v[i-1])
				u = float64((-fil.w / 2) + fil.w*(float64(i)+du)/float64(fil.n-1))

			}
			break
		}
	}

	for i := 0; i < len(fil.cdf_vu[u_i]); i++ {
		if r1 < fil.cdf_vu[u_i][i] {

			// Lerp between stored CDF values to avoid discretisation errors.
			if i == 0 {
				dv := r1 / fil.cdf_vu[u_i][i]
				v = float64((-fil.w / 2)) + (dv / fil.w)
			} else {
				dv := (r1 - fil.cdf_vu[u_i][i-1]) / (fil.cdf_vu[u_i][i] - fil.cdf_vu[u_i][i-1])
				v = float64((-fil.w / 2) + fil.w*(float64(i)+dv)/float64(fil.n-1))

			}

			// These snap to grid points
			//	u = float64((-fil.w / 2) + fil.w*float64(u_i)/float64(fil.n-1))
			//v = float64((-fil.w / 2) + fil.w*float64(i)/float64(fil.n-1))
			return
		}
	}

	return 0, 0
}

func CreateFilterSampler(n int, w float64, f func(x, y float64) float64) (fil *FilterSampler) {
	fil = &FilterSampler{}

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

	fil.p_v = make([]float64, n)

	for j := 0; j < n; j++ {

		for i := 0; i < n; i++ {
			fil.p_v[j] += fil.pdf[j][i]
		}
	}

	// log.Printf("p_v: %v", fil.p_v)

	fil.p_vu = make([][]float64, n)

	for j := 0; j < n; j++ {
		fil.p_vu[j] = make([]float64, n)

		for i := 0; i < n; i++ {
			fil.p_vu[j][i] = fil.pdf[j][i] / fil.p_v[j]
		}
	}

	p := float64(0)
	fil.cdf_v = make([]float64, n)
	for i := 0; i < n; i++ {
		p += fil.p_v[i]
		fil.cdf_v[i] = p
	}

	// log.Printf("cdf_v: %v", fil.cdf_v)

	fil.cdf_vu = make([][]float64, n)
	for j := 0; j < n; j++ {
		fil.cdf_vu[j] = make([]float64, n)

		p := float64(0)

		for i := 0; i < n; i++ {
			p += fil.p_vu[j][i]
			fil.cdf_vu[j][i] = p
		}
	}

	return
	// Finally, select u by cdf_v, then select v from cdf_uv
	// using the appropriate index
}
