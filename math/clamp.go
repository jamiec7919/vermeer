// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// Clamp x to within [min,max]
func Clamp(x, min, max float32) float32 {
	return Max(min, Min(x, max))
}
