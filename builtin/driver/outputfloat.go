// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package driver

import (
	"encoding/binary"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/nodes"
	"os"
)

// OutputFloat is a node which saves the rendered image into a Float file.
type OutputFloat struct {
	NodeDef  core.NodeDef `node:"-"`
	Filename string
}

// Name is a core.Node method.
func (n *OutputFloat) Name() string { return "OutputFloat<>" }

// Def is a core.Node method.
func (n *OutputFloat) Def() core.NodeDef { return n.NodeDef }

// PreRender is a core.Node method.
func (n *OutputFloat) PreRender() error { return nil }

// PostRender is a core.Node method.
func (n *OutputFloat) PostRender() error {
	//w, h := core.FrameMetrics()

	fp, err := os.Create(n.Filename)

	if err != nil {
		return err
	}

	err = binary.Write(fp, binary.LittleEndian, core.FrameBuf())

	return err
}

func init() {
	nodes.Register("OutputFloat", func() (core.Node, error) {
		out := OutputFloat{Filename: "out.float"}

		return &out, nil
	})
}
