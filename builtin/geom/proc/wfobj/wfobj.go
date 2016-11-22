// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wfobj

import (
	//"fmt"
	"github.com/jamiec7919/vermeer/builtin/geom/proc"
	//	"github.com/jamiec7919/vermeer/builtin/geom/polymesh"
	"github.com/jamiec7919/vermeer/core"
	"log"
	"os"
)

type WFObjFile struct {
	Filename string
}

func (f *WFObjFile) Init(p *proc.Proc, datastring string, userdata interface{}) error {
	f.Filename = datastring

	r, err := os.Open(f.Filename)
	defer r.Close()

	if err != nil {
		return err
	}

	mesh, shaders, err := parse(r)

	if err != nil {
		return err
	}

	for _, shader := range shaders {
		mesh.Shader = append(mesh.Shader, shader.MtlName)
		p.Shader = append(p.Shader, shader)
		core.AddNode(shader)
	}

	log.Printf("%v", mesh.Shader)
	mesh.NodeName = f.Filename

	p.Geom = append(p.Geom, mesh)

	if err := mesh.PreRender(); err != nil {
		log.Printf("WFObjFile.Init: %v", err)
	}

	log.Printf("%v", mesh.Bounds(0))
	return nil

}

func create() (proc.Handler, error) {
	mfile := WFObjFile{}

	return &mfile, nil
}

func init() {
	proc.RegisterHandler("wfobj", create)
}
