// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fresnel

// Fresnel model type
type Model int

// Fresnel model constants
const (
	DielectricModel Model = iota
	ConductorModel
)
