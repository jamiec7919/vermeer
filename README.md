# Vermeer Light Tools
[![Go Report Card](https://goreportcard.com/badge/github.com/jamiec7919/vermeer)](https://goreportcard.com/report/github.com/jamiec7919/vermeer)
[![GoDoc](https://godoc.org/github.com/jamiec7919/vermeer?status.svg)](https://godoc.org/github.com/jamiec7919/vermeer)

Vermeer Light Tools is an open source 3D graphics rendering package.  Vermeer provides a ray tracing
core, physically plausible shading and deterministic monte-carlo path tracing all
wrapped up into a production renderer.  

The philosophy behind Vermeer is 'physical as possible but the artist is in charge'.  Pragmatic considerations come before physical realism and any attribute may be pushed or pulled beyond the
realms of reality. If it looks right, it IS right.

Visit the [website](http://www.vermeerlt.com).

This README focuses on the code, refer to the user guide for details on usage.  

Vermeer is written in Go, performance is good but not (yet) on par with optimal C code.  Initially
calling out to C was tried for the inner traversal code however the complexity it added isn't worth the speedup.  Pure Go implementation is the goal, only for very specific libraries there may be C integration - OpenSubDiv or OpenExr might be candidates. 

Installation
------------

To install from source into an existing Go installation, just do:

	go get github.com/jamiec7919/vermeer
	go install github.com/jamiec7919/vermeer

Use
---

Create a .vnf file following the guide, then kick off a render with:

	vermeer <file>.vnf

Contribute
----------

 - Issue Tracker: github.com/jamiec7919/vermeer/issues
 - Source: github.com/jamiec7919/vermeer

Support
-------

If you're having issues or want to discuss, please feel free to contact me [Jamie](jamiec7919@gmail.com).

License
-------

Vermeer Light Tools is released under the BSD license (same as the Go license).  See the LICENSE file
for the full license.

