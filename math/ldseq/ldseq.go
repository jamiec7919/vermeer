// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package ldseq provides functions for sampling based on low-discrepancy scrambled Sobol,
Van der Corput and Larscher Pillichshammer sequences.

Based on 'Efficient Multidimensional Sampling' by Kollig and Keller 2002.  (http://www.uni-kl.de/AG-Heinrich/EMS.pdf)

Implies that can use (0,n,1) net from (0,1) sequence for time & lambda (uncorrelated, stratified over 1 dimension)
And (0,n=2*k,2) nets for 2d lens area (samples per pixel), light source and hemispherical
glossy sampling (and env).  Each dimensional pair can use digital scrambling with a random
integer to decorrelate from the other dimensions that shouldn't be correlated.

Actually suggesting using first n samples from a (0,2) sequence.
For each AA sample x_i take the next b^m samples from the 'light source' or other sequence forming
t,m,s nets for each x_i. This reduces correlation and stratifies across more dimensions. */
package ldseq

// VanDerCorput computes the Van der Corput radical inverse in base 2 with 52 bits precision.
// Returns a float64 in [0,1)
func VanDerCorput(i, scramble uint64) float64 {
	return float64(vanDerCorput(i, scramble)) / float64(uint64(1)<<52)

}

func vanDerCorput(i, scramble uint64) uint64 {
	bits := (i << 32) | (i >> 32)

	bits = ((bits & 0x0000ffff0000ffff) << 16) |
		((bits & 0xffff0000ffff0000) >> 16)

	bits = ((bits & 0x00ff00ff00ff00ff) << 8) |
		((bits & 0xff00ff00ff00ff00) >> 8)

	bits = ((bits & 0x0f0f0f0f0f0f0f0f) << 4) |
		((bits & 0xf0f0f0f0f0f0f0f0) >> 4)

	bits = ((bits & 0x3333333333333333) << 2) |
		((bits & 0xcccccccccccccccc) >> 2)

	bits = ((bits & 0x5555555555555555) << 1) |
		((bits & 0xaaaaaaaaaaaaaaaa) >> 1)

	return (scramble ^ bits) >> (64 - 52) // Account for 52 bits precision.

}

// Sobol calculates the Sobol sequence radical inverse in base 2 with 52 bits precision.
// Returns a float64 in [0,1)
func Sobol(i, scramble uint64) float64 {
	return float64(sobol(i, scramble)) / float64(uint64(1)<<52)
}

func sobol(i, scramble uint64) uint64 {
	r := scramble >> (64 - 52)

	for v := uint64(1) << (52 - 1); i != 0; i >>= 1 {
		if i&1 != 0 {
			r ^= v

		}

		v ^= v >> 1
	}

	return r
}

// LarcherPillichshammer computes Larcher Pillichshammer radical inverse in base 2 with 52 bits precision.
// Returns a float64 in [0,1)
func LarcherPillichshammer(i, scramble uint64) float64 {
	return float64(larcherPillichshammer(i, scramble)) / float64(uint64(1)<<(52))
}

func larcherPillichshammer(i, scramble uint64) uint64 {
	r := scramble >> (64 - 52)

	for v := uint64(1) << (52 - 1); i != 0; i >>= 1 {
		if i&1 != 0 {
			r ^= v

		}

		v |= v >> 1
	}

	return r
}
