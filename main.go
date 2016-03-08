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
	_ "github.com/jamiec7919/vermeer/internal/geom/wfobj"
	_ "github.com/jamiec7919/vermeer/internal/light/disk"
	_ "github.com/jamiec7919/vermeer/internal/material"
	"github.com/jamiec7919/vermeer/preview"
	"log"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

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

	if err := preview.Init(); err != nil {
		log.Printf("Error: preview: %v", err)
		os.Exit(1)
	}

	renderstatus := make(chan error)

	finish := make(chan bool)

	go func() {

		if err := rc.LoadNodeFile(filename); err != nil {
			log.Printf("Error: LoadNodeFile: %v", err)
			renderstatus <- err
			return
		}
		if err := rc.PreRender(); err != nil {
			log.Printf("Error: PreRender: %v", err)
			renderstatus <- err
			return
		}

		if err := rc.Render(finish); err != nil {
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

	preview.Run(rc)

	finish <- true

	<-renderstatus
}
