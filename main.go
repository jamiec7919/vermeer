// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	The vermeer command
*/
package main

import (
	"flag"
	_ "github.com/jamiec7919/vermeer/internal/camera"
	"github.com/jamiec7919/vermeer/internal/core"

	"log"
	"os"
)

func main() {
	flag.Parse()

	filename := "test.vnf"

	if fn := flag.Arg(0); fn != "" {
		filename = fn
	}

	rc := core.NewRenderContext()

	if err := rc.LoadNodeFile(filename); err != nil {
		log.Printf("Error: LoadNodeFile: %v", err)
		os.Exit(1)
	}

	if err := rc.PreRender(); err != nil {
		log.Printf("Error: PreRender: %v", err)
		os.Exit(1)
	}

	if err := rc.Render(); err != nil {
		log.Printf("Error: Render: %v", err)
		os.Exit(1)
	}

	if err := rc.PostRender(); err != nil {
		log.Printf("Error: PostRender: %v", err)
		os.Exit(1)
	}
}
