// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

type MapSampler interface {
	SampleRGB(s, t, ds, dt float32) (out [3]float32)
	SampleScalar(s, t, ds, dt float32) (out float32)
}

type ScalarSampler interface {
	Sample(u, v float32) float32
}

type TripleSampler interface {
	Sample(u, v float32) (r, g, b float32)
}
