// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package driver

import (
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/image"
	_ "github.com/jamiec7919/vermeer/image/hdr"
	"github.com/jamiec7919/vermeer/nodes"
)

// OutputHDR is a node which saves the rendered image intoa Radiance HDR file.
type OutputHDR struct {
	NodeDef  core.NodeDef `node:"-"`
	Filename string
}

// Name is a core.Node method.
func (n *OutputHDR) Name() string { return "OutputHDR<>" }

// Def is a core.Node method.
func (n *OutputHDR) Def() core.NodeDef { return n.NodeDef }

// PreRender is a core.Node method.
func (n *OutputHDR) PreRender() error { return nil }

// PostRender is a core.Node method.
func (n *OutputHDR) PostRender() error {
	i, err := image.NewWriter(n.Filename)

	if err != nil {
		return err
	}

	w, h := core.FrameMetrics()

	spec := image.Spec{
		Width:  w,
		Height: h,
	}

	if err := i.Open(n.Filename, &spec); err != nil {
		return err
	}

	ty := image.TypeDesc{BaseType: image.FLOAT}

	if err := i.WriteImage(ty, core.FrameBuf()); err != nil {
		return err
	}

	i.Close()

	return nil
}

func init() {
	nodes.Register("OutputHDR", func() (core.Node, error) {
		out := OutputHDR{Filename: "out.hdr"}

		return &out, nil
	})
}
