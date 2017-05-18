// Copyright 2017 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// NOTE: These rely on the hardware instructions and don't do special NaN handling
// as standard Go library does.

// Max returns the larger of x or y.
func Max(x, y float32) float32

// Min returns the smaller of x or y.
func Min(x, y float32) float32
