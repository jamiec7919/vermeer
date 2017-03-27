# Vermeer Light Tools
[![Go Report Card](https://goreportcard.com/badge/github.com/jamiec7919/vermeer)](https://goreportcard.com/report/github.com/jamiec7919/vermeer)
[![GoDoc](https://godoc.org/github.com/jamiec7919/vermeer?status.svg)](https://godoc.org/github.com/jamiec7919/vermeer)

Vermeer Light Tools is an open source 3D graphics rendering package.  Vermeer provides a ray tracing
core, physically plausible shading and deterministic monte-carlo path tracing all
wrapped up into a production renderer.  

The philosophy behind Vermeer is 'physical as possible but the artist is in charge'.  Pragmatic considerations come before physical 
realism and any attribute may be pushed or pulled beyond the realms of reality. If it looks right, it IS right.

Visit the [website](http://www.vermeerlt.com).

This README focuses on the code, refer to the user guide for details on usage.  

Vermeer is written in Go, performance is good but not (yet) on par with optimal C code.  Initially
calling out to C was tried for the inner traversal code however the complexity it added isn't worth the speedup.  Pure Go implementation is the goal, only for very specific libraries there may be C integration - OpenSubDiv or OpenExr might be candidates. 

Installation
------------

To install from source into an existing Go installation, just do:

	go install github.com/jamiec7919/vermeer/cmd/vermeer

Use
---

Create a .vnf file following the guide, then kick off a render with:

	vermeer <file>.vnf

Contribute
----------

Vermeer is a work in progress, any and all help would be gratefully received. Some of the things that you
could work on (not an exclusive list!):

 - Bug hunting: Just try running as complex and awkward scenes as possible and report any failures.
 - Add test scenes: Ideally Vermeer would have both a reference test suite plus some complex demo scenes.
 - Improve standard shader/add new shaders:  the default shader needs quite a bit of work to become highly 
   realistic, plus more specialised shaders are welcome.
 - Improve performance of ray tracing:  both low level improvements and code to improve the quality of 
   generated acceleration structures would be useful.
 - Review the maths:  The lighting calculations are spread throughout the code so eyes on to check the
   sampling and integration code would be very useful.
 - Improve spectral handling: there are several improvements to be made in the spectral handling code for
   reflectances.
 - Input and output image formats:  It would be useful to support more formats like EXR and TIFF.
 - Exporters from 3D packages: To get Vermeer into any sort of useful production state it needs to support
   all the major 3D packages.

 - Please report any issues no matter how small: [Issue Tracker](https://github.com/jamiec7919/vermeer/issues)
 - Browse the [Source](https://github.com/jamiec7919/vermeer)

Support
-------

If you're having issues or want to discuss, please feel free to contact me [Jamie](mailto:jamiec7919@gmail.com).

License
-------

Vermeer Light Tools is released under the BSD license (same as the Go license).  See the LICENSE file
for the full license.

