// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ldseq

// Return the index of the frame-th sample falling
// into the square elementary interval (px, py),
// without using look-up tables.
func lookup(m, frame, px, py uint32, scrambleX, scrambleY uint64) uint64 {

	m2 := m << 1
	index := uint64(frame) << m2

	// Note: the delta value only depends on frame
	// and m, thus it can be cached across multiple
	// function calls, if desired.
	delta := uint64(0)
	for c := uint32(0); frame != 0; frame >>= 1 {
		if frame&1 != 0 { // Add flipped column m + c + 1.
			delta ^= vdcSobolMatrices[m-1][c]
		}

		c++
	}

	px ^= uint32(scrambleX >> (64 - m))
	py ^= uint32(scrambleY >> (64 - m))

	b := ((uint64(px) << m) | uint64(py)) ^ delta // flipped b

	for c := uint32(0); b != 0; b >>= 1 {
		if b&1 != 0 { // Add column 2 * m - c.
			index ^= vdcSobolMatricesInv[m-1][c]
		}

		c++
	}
	return index
}

// RasterXY computes the floating-point raster position for the
// frame-th sample falling into the pixel (px, py),
// and return the corresponding sample index.
//
// This function allows efficient stratification across the whole image (not
// just within a pixel) while still allowing per-pixel evaluation of the sample positions.
//
// 2^m must be >= max{width, height}.
func RasterXY(m, frame, px, py uint32, scrambleX, scrambleY uint64) (index uint64, rx, ry float64) {

	index = lookup(m, frame, px, py, scrambleX, scrambleY)

	rx = float64(vanDerCorput(index, scrambleX)) / float64(uint64(1)<<(52-m))
	ry = float64(sobol(index, scrambleY)) / float64(uint64(1)<<(52-m))
	return
}
