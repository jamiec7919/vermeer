// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package fresnel implements various Fresnel models.
*/
package fresnel

// Model is the tyoe of Fresnel model constants.
type Model int

// Fresnel model constants.
const (
	DielectricModel Model = iota
	ConductorModel
)
