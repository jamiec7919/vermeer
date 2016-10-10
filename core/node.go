// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

// NodeDef represents where a node is defined.
type NodeDef struct {
	Where    string // Filename or name of node for auto-generated nodes.
	Col, Row int    // Column and row in file (if applicable)
}

// Node represents a node in the core system.
type Node interface {

	// Name returns the name of the node.
	Name() string

	// Def returns where this node was defined.
	Def() NodeDef

	// PreRender is called after loading scene and before render starts.  Nodes should
	// perform all init and may add other nodes in PreRender.
	PreRender() error

	// PostRender is called after render is complete.
	PostRender() error
}
