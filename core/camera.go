// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

// Camera represents a 3D camera.
type Camera interface {
	Name() string

	// ComputeRay should return a world-space ray.
	ComputeRay(u, v float32, rnd *rand.Rand) (P, D m.Vec3)
}
