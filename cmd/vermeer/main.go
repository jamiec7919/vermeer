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
	"fmt"
	_ "github.com/jamiec7919/vermeer/builtin/camera"
	_ "github.com/jamiec7919/vermeer/builtin/driver"
	_ "github.com/jamiec7919/vermeer/builtin/filter"
	_ "github.com/jamiec7919/vermeer/builtin/geom/polymesh"
	_ "github.com/jamiec7919/vermeer/builtin/geom/proc"
	_ "github.com/jamiec7919/vermeer/builtin/geom/proc/vnf"
	_ "github.com/jamiec7919/vermeer/builtin/geom/proc/wfobj"
	_ "github.com/jamiec7919/vermeer/builtin/light"
	_ "github.com/jamiec7919/vermeer/builtin/misc"
	"github.com/jamiec7919/vermeer/builtin/scene"
	_ "github.com/jamiec7919/vermeer/builtin/shader"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/nodes"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var maxiter = flag.Int("maxiter", -1, "Maximum iterations")
var stats = flag.Bool("stats", false, "stats will be appended to file")
var statsfile = flag.String("statsfile", "stats.txt", "file to append stats to")

func main() {
	flag.Parse()

	debug.SetGCPercent(800)

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

	go http.ListenAndServe(":6060", nil)

	// Capture ctrl-C, finish current iteration and exit.
	c := make(chan os.Signal, 1)

	exit := make(chan bool)

	go func() {
		for range c {
			// sig is a ^C, handle it
			exit <- true
		}
	}()
	//

	core.Init(scene.New())

	if err := nodes.Parse(filename); err != nil {
		log.Printf("Error: LoadNodeFile: %v", err)
		return
	}

	if err := core.PreRender(); err != nil {
		log.Printf("Error: PreRender: %v", err)
		return
	}

	// Capture ctrl-C, finish current iteration and exit.
	signal.Notify(c, os.Interrupt)

	if raystats, err := core.Render(*maxiter, exit); err == nil {

		fmt.Printf("%v\n", raystats)

		if *stats {
			f, err := os.OpenFile(*statsfile, os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				panic(err)
			}

			defer f.Close()

			if _, err := fmt.Fprintln(f, filename, raystats, runtime.Version()); err != nil {
				panic(err)
			}
		}
	} else {
		log.Printf("Error: Render: %v", err)
		return
	}

	if err := core.PostRender(); err != nil {
		log.Printf("Error: PostRender: %v", err)
		return
	}

}
