// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package core provides core interfaces and main render control paths.
*/
package core

var stats RenderStats
var scene Scene
var nodes []Node
var nodeMap map[string]Node
var globals Globals
var filter PixelFilter

var defaultGlobals = Globals{
	NodeDef:       NodeDef{Where: "<auto>"},
	XRes:          1024,
	YRes:          1024,
	MaxGoRoutines: 5,
	MaxIter:       16,
}

// Init initializes the core system with the given Scene.
func Init(s Scene) {
	globals = defaultGlobals
	scene = s
	nodes = nil
	nodeMap = make(map[string]Node)
	stats = RenderStats{}
}

// PreRender is called after all nodes are loaded and calls PreRender on all nodes.
// Nodes may add new nodes so PreRender iterates until no new nodes are created.
func PreRender() error {

	framebuffer = &Framebuffer{globals.XRes, globals.YRes, make([]float32, globals.XRes*globals.YRes*3)}

	// pre and fixup nodes
	// Note that nodes in PreRender may add new nodes, so we must backup and
	// keep track of the existing set so they are only processed once.

	var allnodes []Node

	for nodes != nil {

		_nodes := nodes
		nodes = nil
		allnodes = append(allnodes, _nodes...)

		for _, node := range _nodes {
			if err := node.PreRender(); err != nil {
				return err
			}
		}
	}

	nodes = allnodes

	return scene.PreRender()
}

// PostRender is called on all nodes once Render has returned.
func PostRender() error {
	// post process image
	for _, node := range nodes {
		if err := node.PostRender(); err != nil {
			return err
		}
	}

	return nil
}

// AddNode adds a node to the core.
func AddNode(node Node) {
	nodes = append(nodes, node)
	nodeMap[node.Name()] = node

	switch t := node.(type) {
	case Camera:
		// scene.AddCamera(t)
	case Geom:
		scene.AddGeom(t)
	case Light:
		scene.AddLight(t)
	case PixelFilter:
		filter = t
	case *Globals:
		globals = *t
	}
}

// FindNode finds the node with the given name.
func FindNode(name string) Node {
	node, present := nodeMap[name]

	if present {
		return node
	}

	return nil

}
