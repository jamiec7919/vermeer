// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

// Globals is a node representing the global render settings.
type Globals struct {
	NodeDef       NodeDef  `node:",opt"`
	XRes, YRes    int      `node:",opt"`
	UseProgress   bool     `node:",opt"`
	MaxGoRoutines int      `node:",opt"`
	Camera        string   `node:",opt"`
	MaxIter       int      `node:",opt"`
	Output        []string `node:",opt"` // Output is a string array of AOVs, each string is of format "<AOV name> <format> <filter> <driver>"
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
