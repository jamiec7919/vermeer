// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

// PixelFilter nodes represent sample warp filters for Filter Importance Sampling.
type PixelFilter interface {
	WarpSample(r0, r1 float64) (float64, float64)
}
