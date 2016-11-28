// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnf

import (
	"fmt"
	"github.com/jamiec7919/vermeer/builtin/geom/proc"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/nodes"
	"log"
)

type VNFFile struct {
	Filename string
}

func (f *VNFFile) Init(p *proc.Proc, datastring string, userdata interface{}) error {
	f.Filename = datastring

	sceneNodes, err := nodes.Parse2(datastring)

	if err != nil {
		return err
	}

	for _, node := range sceneNodes {

		if _, ok := node.(core.Shader); ok {
			core.AddNode(node)
		}
	}

	for _, node := range sceneNodes {

		if err != nil {
			log.Printf("VNFFile.Init: %v", err)
		}

		switch t := node.(type) {
		case core.Geom:
			err = node.PreRender()
			p.Geom = append(p.Geom, t)
		case core.Shader:
			p.Shader = append(p.Shader, t)
		default:
			log.Printf("VNFFile.Init: Node %v (%v) is not a Geom or Shader.", node.Name(), node.Def())
		}
	}

	if len(p.Geom) == 0 {
		return fmt.Errorf("GeomProc: %v no geoms.", p.Name())
	}

	return nil
}

func create() (proc.Handler, error) {
	mfile := VNFFile{}

	return &mfile, nil
}

func init() {
	proc.RegisterHandler("vnf", create)
}
