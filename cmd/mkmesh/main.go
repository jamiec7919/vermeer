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

func loadMesh(filename string) error {

	f, err := os.Open(filename)

	if err != nil {
		return err
	}

	defer f.Close()

	meshes, err := wfobj.Load(f)

	fout, err := os.Create("test.vnf")
	defer fout.Close()
	for i, m := range meshes {
		m.WriteNodes(fout, filename, fmt.Sprintf("mesh%v", i))
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
