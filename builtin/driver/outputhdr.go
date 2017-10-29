// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package driver

import (
	"fmt"
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/image"
	_ "github.com/jamiec7919/vermeer/image/hdr"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/nodes"
	"log"
)

// OutputHDR is a node which saves the rendered image intoa Radiance HDR file.
type OutputHDR struct {
	NodeDef  core.NodeDef `node:"-"`
	NodeName string       `node:"Name"`
	Filename string
}

// Assert that node implements relevant interfaces.
var _ core.Node = (*OutputHDR)(nil)
var _ core.Driver = (*OutputHDR)(nil)

// Name is a core.Node method.
func (n *OutputHDR) Name() string { return n.NodeName }

// Def is a core.Node method.
func (n *OutputHDR) Def() core.NodeDef { return n.NodeDef }

// PreRender is a core.Node method.
func (n *OutputHDR) PreRender() error { return nil }

// PostRender is a core.Node method.
func (n *OutputHDR) PostRender() error {
	/*
		i, err := image.NewWriter(n.Filename)

		if err != nil {
			return err
		}

		w, h := core.FrameMetrics()

		spec := image.Spec{
			Width:  w,
			Height: h,
		}

		if err := i.Open(n.Filename, &spec); err != nil {
			return err
		}

		ty := image.TypeDesc{BaseType: image.FLOAT}

		// framebuffer is XYZ
		rgb := make([]float32, w*h*3)

		for k := 0; k < w*h*3; k += 3 {
			xyz := [3]float32{core.FrameBuf()[k+0], core.FrameBuf()[k+1], core.FrameBuf()[k+2]}
			xyz = colour.ChromaticAdjust(colour.XYZScalingEToD65, xyz)

			//norm := 1 / (xyz[0] + xyz[1] + xyz[2])
			// I think converting to xyY then sRGB then scaling by the luminance might result in slightly better behaviour
			// but realistically we should either output XYZ or do proper tone mapping (currently just relying on RGBE/radiance)

			//if norm < m.Float32Max {
			//	col := colour.SRGB.XYZToRGB(xyz[0]*norm, xyz[1]*norm, xyz[2]*norm)
			//	rgb[k+0] = col[0] / norm
			//	rgb[k+1] = col[1] / norm
			//	rgb[k+2] = col[2] / norm
			//}

			//xyz = colour.ChromaticAdjust(colour.BradfordD65ToD50, xyz)

			col := colour.SRGB.XYZToRGB(xyz[0], xyz[1], xyz[2])

			// Fix gamut issue where components are < 0
			for k := range col {
				col[k] = m.Max(0, col[k])
			}

			rgb[k+0] = col[0]
			rgb[k+1] = col[1]
			rgb[k+2] = col[2]
		}

		if err := i.WriteImage(ty, rgb); err != nil {
			return err
		}

		i.Close()

		log.Printf("Wrote %s", n.Filename)
	*/
	return nil
}

// Write implements core.Driver
func (n *OutputHDR) Write(fb *core.Framebuffer, aovs []string) error {
	i, err := image.NewWriter(n.Filename)

	if err != nil {
		return err
	}

	w := fb.Width()
	h := fb.Height()

	spec := image.Spec{
		Width:  w,
		Height: h,
	}

	if err := i.Open(n.Filename, &spec); err != nil {
		return err
	}

	ty := image.TypeDesc{BaseType: image.FLOAT}

	xyz := fb.AOV("RGB")

	if xyz == nil {
		return fmt.Errorf("OutputHDR.Write: no AOV \"RGB\"")
	}

	var buf []float32

	switch t := xyz.(type) {
	case []float32:
		buf = t
	default:
		return fmt.Errorf("OutputHDR.Write: AOV \"RGB\" is wrong format")
	}

	// framebuffer is XYZ
	rgb := make([]float32, w*h*3)

	for k := 0; k < w*h*3; k += 3 {
		xyz := [3]float32{buf[k+0], buf[k+1], buf[k+2]}
		xyz = colour.ChromaticAdjust(colour.XYZScalingEToD65, xyz)

		//norm := 1 / (xyz[0] + xyz[1] + xyz[2])
		// I think converting to xyY then sRGB then scaling by the luminance might result in slightly better behaviour
		// but realistically we should either output XYZ or do proper tone mapping (currently just relying on RGBE/radiance)

		//if norm < m.Float32Max {
		//	col := colour.SRGB.XYZToRGB(xyz[0]*norm, xyz[1]*norm, xyz[2]*norm)
		//	rgb[k+0] = col[0] / norm
		//	rgb[k+1] = col[1] / norm
		//	rgb[k+2] = col[2] / norm
		//}

		//xyz = colour.ChromaticAdjust(colour.BradfordD65ToD50, xyz)

		col := colour.SRGB.XYZToRGB(xyz[0], xyz[1], xyz[2])

		// Fix gamut issue where components are < 0
		for k := range col {
			col[k] = m.Max(0, col[k])
		}

		rgb[k+0] = col[0]
		rgb[k+1] = col[1]
		rgb[k+2] = col[2]
	}

	if err := i.WriteImage(ty, rgb); err != nil {
		return err
	}

	i.Close()

	log.Printf("Wrote %s", n.Filename)
	return nil
}

func init() {
	nodes.Register("OutputHDR", func() (core.Node, error) {
		out := OutputHDR{Filename: "out.hdr"}

		return &out, nil
	})
}
