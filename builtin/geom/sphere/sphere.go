// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sphere

import (
	"fmt"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/nodes"
)

// Sphere geoms are only used for spherical light sources in Vermeer.
type Sphere struct {
	NodeDef  core.NodeDef `node:"-"`
	NodeName string       `node:"Name"`
	RayBias  float32      `node:",opt"`

	P      m.Vec3
	Radius float32
	Shader string

	shader core.Shader

	bounds       m.BoundingBox
	motionBounds []m.BoundingBox
}

// Assert that Sphere implements important interfaces.
var _ core.Node = (*Sphere)(nil)
var _ core.Geom = (*Sphere)(nil)

// Name is a core.Node method.
func (sphere *Sphere) Name() string { return sphere.NodeName }

// Def is a core.Node method.
func (sphere *Sphere) Def() core.NodeDef { return sphere.NodeDef }

// PreRender is a core.Node method.
func (sphere *Sphere) PreRender() error {

	if s := core.FindNode(sphere.Shader); s != nil {
		shader, ok := s.(core.Shader)

		if !ok {
			return fmt.Errorf("Unable to find shader %v", sphere.Shader)
		}

		sphere.shader = shader
	} else {
		return fmt.Errorf("Unable to find node (shader %v)", sphere.Shader)

	}

	return nil
}

// PostRender is a core.Node method.
func (sphere *Sphere) PostRender() error { return nil }

// MotionKeys returns the number of motion keys.
func (sphere *Sphere) MotionKeys() int {
	return 1
}

// Bounds implements core.Geom.
func (sphere *Sphere) Bounds(time float32) (b m.BoundingBox) {

	for k := range b.Bounds[0] {
		b.Bounds[0][k] = sphere.P.Elt(k) - sphere.Radius
	}
	for k := range b.Bounds[1] {
		b.Bounds[1][k] = sphere.P.Elt(k) + sphere.Radius
	}
	return
}

func create() (core.Node, error) {

	return &Sphere{Radius: 1}, nil
}

func init() {
	nodes.Register("Sphere", create)
}
