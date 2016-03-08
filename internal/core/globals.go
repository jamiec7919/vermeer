// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

type Globals struct {
	XRes, YRes    int
	UseProgress   bool
	MaxGoRoutines int
}

func init() {
	RegisterType("Globals", func(rc *RenderContext, params Params) (interface{}, error) {
		return nil, params.Unmarshal(&rc.globals)
	})
}
