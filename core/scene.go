// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

// Scene represents the top level scene.
type Scene interface {
	// Trace returns true if ray hits something. Usually this is the first hit along the ray
	// unless ray.Type &= RayTypeShadow.
	Trace(*Ray, *ShaderContext) bool

	// LightsPrepare initializes the context with a potential set of lights.
	LightsPrepare(*ShaderContext)

	// AddGeom adds the geom to the scene.
	AddGeom(Geom) error

	// AddLight adds the light to the scene.
	AddLight(Light) error

	// PreRender is called after all other nodes PreRender.
	PreRender() error
}
