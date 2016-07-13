// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
)

// Version constants.
const (
	VersionMajor   = 0
	VersionMinor   = 1
	VersionPatch   = 0
	VersionRelease = "alpha"
)

// VersionString returns a string representing the current software version.
func VersionString() string {
	if VersionRelease == "" {
		return fmt.Sprintf("%v.%v.%v", VersionMajor, VersionMinor, VersionPatch)
	} else {
		return fmt.Sprintf("%v.%v.%v-%v", VersionMajor, VersionMinor, VersionPatch, VersionRelease)
	}
}
