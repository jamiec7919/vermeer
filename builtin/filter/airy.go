// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filter

import (
	"github.com/jamiec7919/vermeer/core"
	"math"
)

// AiryFilter implements the Airy filter model.
type AiryFilter struct {
	NodeDef  core.NodeDef `node:"-"`
	NodeName string       `node:"Name"`
	Width    float32      `node:",opt"`
	Res      int          `node:",opt"`
	Peak     float32      `node:",opt"`

	sampler *Sampler
}

// WarpSample implements core.PixelFilter.
func (f *AiryFilter) WarpSample(r0, r1 float64) (u float64, v float64) {
	return f.sampler.WarpSample(r0, r1)
}

// Name is a core.Node method.
func (f *AiryFilter) Name() string { return f.NodeName }

// Def is a core.Node method.
func (f *AiryFilter) Def() core.NodeDef { return f.NodeDef }

// BesselJ1 implements the Bessel function of first kind and order.  I realise
// this is actually implemented in the std lib too!
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

	f.sampler = CreateSampler(f.Res, float64(f.Width), airy)

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
