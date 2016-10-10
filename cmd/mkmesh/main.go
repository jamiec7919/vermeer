// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
mkmesh command converts meshes from various formats into Polymesh nodes suitable for rendering with
Vermeer.

Execute as:

	mkmesh <file.obj> <out.vnf

*/
package main

import (
	"flag"
	"fmt"
	"github.com/jamiec7919/vermeer/cmd/mkmesh/wfobj"
	"os"
)

var output = flag.String("o", "test.vnf", "Output filename")

func loadMesh(filename string) error {

	f, err := os.Open(filename)

	if err != nil {
		return err
	}

	defer f.Close()

	meshes, shaders, err := wfobj.Load(f)

	fout, err := os.Create(*output)
	defer fout.Close()

	for i, m := range meshes {
		m.WriteNodes(fout, filename, fmt.Sprintf("mesh%v", i))
	}

	for _, m := range shaders {
		m.Write(fout, filename)
	}

	return err

}

func main() {
	flag.Parse()

	filename := ""

	if fn := flag.Arg(0); fn != "" {
		filename = fn
	}

	loadMesh(filename)

}
