// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	m "github.com/jamiec7919/vermeer/math"
	"math/rand"
)

type Camera interface {
	Name() string
	ComputeRay(u, v float32, rnd *rand.Rand) (P, D m.Vec3)
}
