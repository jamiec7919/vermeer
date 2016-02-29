// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// Sqrt returns the square root of x.
// This uses SSE4 instruction only. (AMD64)
func Sqrt(x float32) float32
