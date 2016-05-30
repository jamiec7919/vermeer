// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	The vermeer command.

	Execute as:

	  vermeer [-maxiter=n] [-cpuprofile=filename.prof] <file.vnf>
*/
package main

import (
	"flag"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/nodes"
	"github.com/jamiec7919/vermeer/preview"
	"log"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var maxiter = flag.Int("maxiter", -1, "Maximum iterations")

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		log.Printf("CPU profile: %v", *cpuprofile)
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	filename := "test.vnf"

	if fn := flag.Arg(0); fn != "" {
		filename = fn
	}

	rc := core.NewRenderContext()

	pview, err := preview.Init()

	if err != nil {
		log.Printf("Error: preview: %v", err)
		os.Exit(1)
	}

	rc.StartPreview(pview)

	renderstatus := make(chan error)

	go func() {
		defer pview.Close()

		if err := nodes.Parse(rc, filename); err != nil {
			log.Printf("Error: LoadNodeFile: %v", err)
			renderstatus <- err
			return
		}
		if err := rc.PreRender(); err != nil {
			log.Printf("Error: PreRender: %v", err)
			renderstatus <- err
			return
		}

		if err := rc.Render(*maxiter); err != nil {
			log.Printf("Error: Render: %v", err)
			renderstatus <- err
			return
		}

		if err := rc.PostRender(); err != nil {
			log.Printf("Error: PostRender: %v", err)
			renderstatus <- err
			return
		}
		renderstatus <- nil
	}()

	pview.Run() // This blocks until window is closed

	rc.Finish() // If render is still going we finish it

	<-renderstatus
}
