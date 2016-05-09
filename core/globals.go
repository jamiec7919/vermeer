// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

type Globals struct {
	XRes, YRes    int
	UseProgress   bool
	MaxGoRoutines int
}

func (g *Globals) Name() string                    { return "<globals>" }
func (g *Globals) PreRender(*RenderContext) error  { return nil }
func (g *Globals) PostRender(*RenderContext) error { return nil }
