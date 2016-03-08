// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/jamiec7919/vermeer/image"
)

type OutputHDR struct {
	Filename string
}

func (n *OutputHDR) Name() string                      { return "OutputHDR<>" }
func (n *OutputHDR) PreRender(rc *RenderContext) error { return nil }

func (n *OutputHDR) PostRender(rc *RenderContext) error {
	i, err := image.NewWriter(n.Filename)

	if err != nil {
		return err
	}

	spec := image.Spec{
		Width:  rc.globals.XRes,
		Height: rc.globals.YRes,
	}

	if err := i.Open(n.Filename, &spec); err != nil {
		return err
	}

	ty := image.TypeDesc{BaseType: image.FLOAT}

	if err := i.WriteImage(ty, rc.imgbuf); err != nil {
		return err
	}

	i.Close()

	return nil
}

func init() {
	RegisterType("OutputHDR", func(rc *RenderContext, params Params) (interface{}, error) {
		out := OutputHDR{"out.hdr"}

		params.Unmarshal(&out)

		return &out, nil
	})
}
