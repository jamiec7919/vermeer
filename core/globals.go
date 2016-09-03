// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

// Globals is a node representing the global render settings.
type Globals struct {
	NodeDef       NodeDef
	XRes, YRes    int
	UseProgress   bool
	MaxGoRoutines int
	Camera        string
	MaxIter       int
	Output        string
}

var _ Node = (*Globals)(nil)

// Name is a node method.
func (g *Globals) Name() string { return "<globals>" }

// Def is a node method.
func (g *Globals) Def() NodeDef { return g.NodeDef }

// PreRender is a node method.
func (g *Globals) PreRender() error { return nil }

// PostRender is a node method.
func (g *Globals) PostRender() error { return nil }
