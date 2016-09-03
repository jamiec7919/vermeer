// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package param

import (
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
)

// Float32Uniform is used by nodes for parseable parameters. This represents a float32 param
// which may be constant, from a texture or something more complex.
type Float32Uniform interface {
	Float32(sg *core.ShaderContext) float32
}

// RGBUniform is used by nodes for parseable parameters. This represents an RGB colour param
// which may be constant, from a texture or something more complex.
type RGBUniform interface {
	RGB(sg *core.ShaderContext) colour.RGB
}
