// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package material provides the default shader(s) for Vermeer.  (should rename).

This package is in heavy development so documentation somewhat sketchy. */
package material

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/nodes"
)

// Debug is the default surface shader.
type Debug struct {
	MtlName string `node:"Name"`
	id      int32  // This should only be assinged by RenderContext

	Sides  int           // One or two sided
	Colour core.RGBParam // Colour parameter
}

// Assert that Debug satisfies important interfaces.
var _ core.Node = (*Debug)(nil)
var _ core.Material = (*Debug)(nil)

// Name is a core.Node method.
func (mtl *Debug) Name() string { return mtl.MtlName }

// PreRender is a core.Node method.
func (mtl *Debug) PreRender(rc *core.RenderContext) error {
	return nil
}

// PostRender is a core.Node method.
func (mtl *Debug) PostRender(rc *core.RenderContext) error { return nil }

// ID is a core.Debug method.
//
// Deprecated?:
func (mtl *Debug) ID() int32 {
	return mtl.id
}

// SetID is a core.Debug method.
// Deprecated?:
func (mtl *Debug) SetID(id int32) {
	mtl.id = id
}

// Eval implements core.Material.  Performs all shading for the surface point in sg.  May trace
// rays and shadow rays.
func (mtl *Debug) Eval(sg *core.ShaderGlobals) {
	sg.OutRGB = mtl.Colour.RGB(sg)
}

// HasBumpMap implements core.Material.
func (mtl *Debug) HasBumpMap() bool { return false }

// Emission returns the RGB emission for the given direction.
func (mtl *Debug) Emission(sg *core.ShaderGlobals, omegaO m.Vec3) colour.RGB {
	return colour.RGB{}
}

func init() {
	nodes.Register("MaterialDebug", func() (core.Node, error) {

		mtl := &Debug{}
		return mtl, nil
	})
}
