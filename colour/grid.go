// Copyright 2017 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package colour

import (
	m "github.com/jamiec7919/vermeer/math"
)

/*
  Physically Meaningful Rendering using Tristimulus Colours
  https://cg.ivd.kit.edu/spectrum.php

  Unashmedly ported from the supplemental code.  The basic idea is to sample the x-y colour space (using unity brightness plane X+Y+Z=1)
  with a regular grid plus some triangles for the edges (as it's an odd shape).  Each sample consists of a spectrum computed by a minimization
  procedure (we should replicate this eventually from the Python code, would be an interesting test of Go).  When we want to
  compute a new colour then we get the x-y coords and linearly interpolate the spectra from the nearby grid points.
*/

/*
 * Evaluate the spectrum for xyz at the given wavelength.
 */
func spectrumXYZToP(lambda float32, xyz [3]float32) float32 {
	//assert(lambda >= spectrum_sample_min)
	//assert(lambda <= spectrum_sample_max)

	var xyY [3]float32
	var uv [2]float32

	norm := 1.0 / (xyz[0] + xyz[1] + xyz[2])

	if !(norm < m.Float32Max) {
		return 0.0
	}

	// convert to xy chromaticities
	xyY[0] = xyz[0] * norm
	xyY[1] = xyz[1] * norm
	xyY[2] = xyz[1]

	// rotate to align with grid
	tmp := [2]float32{xyY[0], xyY[1]}
	uv = spectrumxyTouv(tmp)

	if uv[0] < 0.0 || uv[0] >= spectrumGridWidth ||
		uv[1] < 0.0 || uv[1] >= spectrumGridHeight {
		return 0.0
	}

	uvi := [2]int{int(uv[0]), int(uv[1])}

	//  assert(uvi[0] < spectrum_grid_width);
	//  assert(uvi[1] < spectrum_grid_height);

	cellIdx := uvi[0] + spectrumGridWidth*uvi[1]
	//assert(cell_idx < spectrum_grid_width*spectrum_grid_height);
	//assert(cell_idx >= 0);

	cell := &spectrumGrid[cellIdx]
	inside := cell.inside
	idx := &cell.idx
	num := int(cell.num_points)

	// get linearly interpolated spectral power for the corner vertices:
	var p [6]float32
	// this clamping is only necessary if lambda is not sure to be >= spectrum_sample_min and <= spectrum_sample_max:
	sb := //fminf(spectrum_num_samples-1e-4, fmaxf(0.0f,
		(lambda - spectrumSampleMin) / (spectrumSampleMax - spectrumSampleMin) * (spectrumNumSamples - 1) //));
	//assert(sb >= 0.f);
	//assert(sb <= spectrum_num_samples);

	sb0 := int(sb)
	sb1 := sb0 + 1

	if sb1 >= spectrumNumSamples {
		sb1 = spectrumNumSamples - 1
	}

	sbf := sb - m.Floor(sb)

	for i := 0; i < num; i++ {
		//assert(idx[i] >= 0);
		//assert(sb0 < spectrum_num_samples);
		//assert(sb1 < spectrum_num_samples);

		p[i] = spectrumDataPoints[idx[i]].spectrum[sb0]*(1.0-sbf) + spectrumDataPoints[idx[i]].spectrum[sb1]*sbf
	}

	var interpolated_p float32

	if inside != 0 { // fast path for normal inner quads:
		uv[0] -= float32(uvi[0])
		uv[1] -= float32(uvi[1])
		//    assert(uv[0] >= 0 && uv[0] <= 1.f);
		//   assert(uv[1] >= 0 && uv[1] <= 1.f);

		// the layout of the vertices in the quad is:
		//  2  3
		//  0  1
		interpolated_p =
			p[0]*(1.0-uv[0])*(1.0-uv[1]) + p[2]*(1.0-uv[0])*uv[1] +
				p[3]*uv[0]*uv[1] + p[1]*uv[0]*(1.0-uv[1])
	} else {
		// need to go through triangulation :(
		// we get the indices in such an order that they form a triangle fan around idx[0].
		// compute barycentric coordinates of our xy* point for all triangles in the fan:
		ex := uv[0] - spectrumDataPoints[idx[0]].uv[0]
		ey := uv[1] - spectrumDataPoints[idx[0]].uv[1]
		e0x := spectrumDataPoints[idx[1]].uv[0] - spectrumDataPoints[idx[0]].uv[0]
		e0y := spectrumDataPoints[idx[1]].uv[1] - spectrumDataPoints[idx[0]].uv[1]
		uu := e0x*ey - ex*e0y

		for i := 0; i < int(num)-1; i++ {
			var e1x, e1y float32

			if i == int(num)-2 { // close the circle
				e1x = spectrumDataPoints[idx[1]].uv[0] - spectrumDataPoints[idx[0]].uv[0]
				e1y = spectrumDataPoints[idx[1]].uv[1] - spectrumDataPoints[idx[0]].uv[1]
			} else {
				e1x = spectrumDataPoints[idx[i+2]].uv[0] - spectrumDataPoints[idx[0]].uv[0]
				e1y = spectrumDataPoints[idx[i+2]].uv[1] - spectrumDataPoints[idx[0]].uv[1]
			}

			vv := ex*e1y - e1x*ey
			// TODO: with some sign magic, this division could be deferred to the last iteration!
			area := e0x*e1y - e1x*e0y
			// normalise
			u := uu / area
			v := vv / area
			w := 1.0 - u - v
			// outside spectral locus (quantized version at least) or outside grid
			if u < 0.0 || v < 0.0 || w < 0.0 {
				uu = -vv
				e0x = e1x
				e0y = e1y
				continue
			}

			// This seems to be the triangle we've been looking for.
			otherIdx := i + 2

			if i == num-2 {
				otherIdx = 1
			}

			//interpolated_p = p[0] * w + p[i+1] * v + p[(i == num-2) ? 1 : (i+2)] * u
			interpolated_p = p[0]*w + p[i+1]*v + p[otherIdx]*u
			break
		}
	}

	// now we have a spectrum which corresponds to the xy chromaticities of the input. need to scale according to the
	// input brightness X+Y+Z now:
	return interpolated_p / norm

}
