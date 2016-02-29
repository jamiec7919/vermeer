// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import "math"

/*
	Float16 (IEEE binary16, sign bit, 5 bit exponent, 10 bit mantissa)
	Go implementation of article:
	http://www.mathworks.com/matlabcentral/fileexchange/23173-ieee-754r-half-precision-floating-point-converter
	(BSD License)
*/

type Float16 uint16

func Float32ToFloat16(xe float32) Float16 {
	x := math.Float32bits(xe)

	if x&0x7FFFFFFF == 0 { // Signed zero
		return Float16(x >> 16) // return the signed zero
	} else {
		xs := x & 0x80000000 // Pick off sign bit
		xe := x & 0x7F800000 // Pick off exponent bits
		xm := x & 0x007FFFFF // Pick off mantissa bits

		if xe == 0 { // Denormal will underflow, return a signed zero
			return Float16(xs >> 16)
		} else if xe == 0x7F800000 { // Inf or NaN (all exponent bits are set)
			if xm == 0 { // If mantissa is zero...
				return Float16((xs >> 16) | 0x7C00) // Signed Inf
			} else {
				return Float16(0xFE00) // NaN, only 1st mantissa bit set
			}
		} else { // Normalized number
			hs := Float16(xs >> 16)              // Sign bit
			hes := (uint32(xe >> 23)) - 127 + 15 // Exponent unbias the single, then bias the halfp
			if hes >= 0x1F {                     // Overflow
				return Float16((xs >> 16) | 0x7C00) // Signed Inf
			} else if hes <= 0 { // Underflow

				var hm Float16

				if (14 - hes) > 24 { // Mantissa shifted all the way off & no rounding possibility
					hm = Float16(0) // Set mantissa to zero
				} else {
					xm |= 0x00800000                    // Add the hidden leading bit
					hm = Float16(xm >> (14 - hes))      // Mantissa
					if (xm>>(13-hes))&0x00000001 != 0 { // Check for rounding
						hm += Float16(1) // Round, might overflow into exp bit, but this is OK
					}
				}
				return (hs | hm) // Combine sign bit and mantissa bits, biased exponent is zero
			} else {
				he := Float16(hes << 10) // Exponent
				hm := Float16(xm >> 13)  // Mantissa
				if xm&0x00001000 != 0 {  // Check for rounding
					return (hs | he | hm) + Float16(1) // Round, might overflow to inf, this is OK
				} else {
					return (hs | he | hm) // No rounding
				}
			}
		}
	}
}

func Float16ToFloat32(h Float16) float32 {
	var x uint32

	if h&0x7FFF == 0 { // Signed zero
		x = uint32(h) << 16 // Return the signed zero
	} else { // Not zero
		hs := h & 0x8000 // Pick off sign bit
		he := h & 0x7C00 // Pick off exponent bits
		hm := h & 0x03FF // Pick off mantissa bits

		if he == 0 { // Denormal will convert to normalized
			e := 0 // The following loop figures out how much extra to adjust the exponent
			hm <<= 1

			for hm&0x0400 == 0 {
				e++
				hm <<= 1
			} // Shift until leading bit overflows into exponent bit
			xs := uint32(hs) << 16                // Sign bit
			xes := (int(he >> 10)) - 15 + 127 - e // Exponent unbias the halfp, then bias the single
			xe := uint32(xes << 23)               // Exponent
			xm := uint32(hm&0x03FF) << 13         // Mantissa
			x = (xs | xe | xm)                    // Combine sign bit, exponent bits, and mantissa bits

		} else if he == 0x7C00 { // Inf or NaN (all the exponent bits are set)
			if hm == 0 { // If mantissa is zero ...
				x = (uint32(hs) << 16) | 0x7F800000 // Signed Inf
			} else {
				x = uint32(0xFFC00000) // NaN, only 1st mantissa bit set
			}
		} else { // Normalized number
			xs := uint32(hs) << 16        // Sign bit
			xes := int(he>>10) - 15 + 127 // Exponent unbias the halfp, then bias the single
			xe := uint32(xes << 23)       // Exponent
			xm := uint32(hm) << 13        // Mantissa
			x = (xs | xe | xm)            // Combine sign bit, exponent bits, and mantissa bits
		}
	}

	return math.Float32frombits(x)
}
