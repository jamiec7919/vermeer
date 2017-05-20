// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package shader

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/core/param"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/nodes"
)

// Debug is the default surface shader.
type Debug struct {
	NodeDef core.NodeDef `node:"-"`
	MtlName string       `node:"Name"`

	Sides  int              `node:",opt"` // One or two sided
	Colour param.RGBUniform // Colour parameter
}

// Assert that Debug satisfies important interfaces.
var _ core.Node = (*Debug)(nil)
var _ core.Shader = (*Debug)(nil)

// Name is a core.Node method.
func (sh *Debug) Name() string { return sh.MtlName }

// Def is a core.Node method.
func (sh *Debug) Def() core.NodeDef { return sh.NodeDef }

// PreRender is a core.Node method.
func (sh *Debug) PreRender() error {
	return nil
}

// PostRender is a core.Node method.
func (sh *Debug) PostRender() error { return nil }

// Eval implements core.Shader.  Performs all shading for the surface point in sg.  May trace
// rays and shadow rays.
func (sh *Debug) Eval(sg *core.ShaderContext) bool {
	//sg.OutRGB = sh.Colour.RGB(sg)
	sg.Out = sh.Colour.RGB(sg)

	return true
}

// EvalEmission implements core.Shader.
func (sh *Debug) EvalEmission(sg *core.ShaderContext, omegaO m.Vec3) colour.Spectrum {
	return colour.Spectrum{Lambda: sg.Lambda}
}

func init() {
	nodes.Register("DebugShader", func() (core.Node, error) {

		sh := &Debug{}
		return sh, nil
	})
}
