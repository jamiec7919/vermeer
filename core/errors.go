// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"errors"
)

// Common errors.
var (
	ErrNodeAlreadyRegistered = errors.New("Node type already registered.")
	ErrNodeNotRegistered     = errors.New("Node type not registered")
	ErrNotNode               = errors.New("Object is not Node")

	ErrNoSample = errors.New("No sample")

	ErrNoCamera = errors.New("No camera")
)
