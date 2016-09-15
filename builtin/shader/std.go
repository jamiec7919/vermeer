// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package material provides the default shader(s) for Vermeer.  (should rename).

This package is in heavy development so documentation somewhat sketchy. */
package shader

import (
	"github.com/jamiec7919/vermeer/builtin/shader/bsdf"
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/core/param"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/nodes"
)

// ShaderStd is the default surface shader.
type ShaderStd struct {
	NodeDef core.NodeDef `node:"-"`
	MtlName string       `node:"Name"`

	EmissionColour   param.RGBUniform     `node:",opt"`
	EmissionStrength param.Float32Uniform `node:",opt"`

	Sides         int              `node:",opt"` // One or two sided
	DiffuseColour param.RGBUniform `node:",opt"` // Colour parameter
}

// Assert that ShaderStd satisfies important interfaces.
var _ core.Node = (*ShaderStd)(nil)
var _ core.Shader = (*ShaderStd)(nil)

// Name is a core.Node method.
func (sh *ShaderStd) Name() string { return sh.MtlName }

// Def is a core.Node method.
func (sh *ShaderStd) Def() core.NodeDef { return sh.NodeDef }

// PreRender is a core.Node method.
func (sh *ShaderStd) PreRender() error {
	return nil
}

// PostRender is a core.Node method.
func (sh *ShaderStd) PostRender() error { return nil }

// Eval implements core.Shader.  Performs all shading for the surface point in sg.  May trace
// rays and shadow rays.
func (sh *ShaderStd) Eval(sg *core.ShaderContext) {

	// Construct a tangent space
	V := m.Vec3Cross(sg.N, sg.DdPdu)

	if m.Vec3Length2(V) < 0.1 {
		V = m.Vec3Cross(sg.N, sg.DdPdv)
	}
	V = m.Vec3Normalize(V)
	U := m.Vec3Cross(sg.N, V)

	//diffBrdf := bsdf.NewOrenNayar(sg.Lambda, m.Vec3Neg(sg.Rd), 0.4, U, V, sg.N)
	diffBrdf := bsdf.NewLambert(sg.Lambda, m.Vec3Neg(sg.Rd), U, V, sg.N)

	var diffContrib colour.RGB

	sg.LightsPrepare()

	var diffColour colour.RGB

	if sh.DiffuseColour != nil {
		diffColour = sh.DiffuseColour.RGB(sg)
	}

	for sg.LightsGetSample() {

		if sg.Lp.DiffuseShadeMult() > 0.0 {

			// In this example the brdf passed is an interface
			// allowing sampling, pdf and bsdf eval
			col := sg.EvaluateLightSample(diffBrdf)
			col.Mul(diffColour)
			diffContrib.Add(col)
		}

	}

	contrib := colour.RGB{}

	emissContrib := sh.EvalEmission(sg, m.Vec3Neg(sg.Rd))

	contrib.Add(emissContrib)
	contrib.Add(diffContrib)

	sg.OutRGB = contrib
}

// EvalEmission implements core.Shader.
func (sh *ShaderStd) EvalEmission(sg *core.ShaderContext, omegaO m.Vec3) colour.RGB {

	var emissColour colour.RGB
	var emissStrength float32 = 0

	if sh.EmissionColour != nil {
		emissColour = sh.EmissionColour.RGB(sg)
	}

	if sh.EmissionStrength != nil {
		emissStrength = sh.EmissionStrength.Float32(sg)
	} else {
		return colour.RGB{}
	}

	emissColour.Scale(emissStrength)
	return emissColour
}

func init() {
	nodes.Register("ShaderStd", func() (core.Node, error) {

		sh := &ShaderStd{}
		return sh, nil
	})
}
