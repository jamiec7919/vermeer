// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

// Camera represents a 3D camera.
type Camera interface {
	// ComputeRay should return a world-space ray within the given pixel.
	ComputeRay(sc *ShaderContext, lensU, lensV float64, ray *Ray)
}
