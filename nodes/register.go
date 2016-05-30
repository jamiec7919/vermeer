// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nodes

import (
	"errors"
	"github.com/jamiec7919/vermeer/core"
)

var nodeTypes = map[string]func() (core.Node, error){}

func createNode(name string) (core.Node, error) {

	create, present := nodeTypes[name]

	if present {
		return create()
	}

	return nil, errors.New("Node type " + name + " not registered.")
}

// Register is called by library nodes to register their names.
func Register(name string, create func() (core.Node, error)) error {
	if nodeTypes[name] == nil {
		nodeTypes[name] = create
		return nil
	}

	return errors.New("Node type " + name + " already registered.")
}
