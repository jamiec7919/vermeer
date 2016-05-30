// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

// Globals is a node representing the global render settings.
type Globals struct {
	XRes, YRes    int
	UseProgress   bool
	MaxGoRoutines int
}

// Name is a node method.
func (g *Globals) Name() string { return "<globals>" }

// PreRender is a node method.
func (g *Globals) PreRender(*RenderContext) error { return nil }

// PostRender is a node method.
func (g *Globals) PostRender(*RenderContext) error { return nil }
