// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wfobj

import (
	"bufio"
	"github.com/jamiec7919/vermeer/internal/core"
	"github.com/jamiec7919/vermeer/material"
	"github.com/jamiec7919/vermeer/material/bsdf"
	"github.com/jamiec7919/vermeer/material/edf"
	"os"
	"strconv"
	"strings"
)

func ParseMtlLib(rc *core.RenderContext, filename string) error {
	fin, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fin.Close()

	//var mtlid int
	scanner := bufio.NewScanner(fin)
	// bytes := make([]byte, DefaultBufferSize)

	var mtl *material.Material

	for scanner.Scan() {
		line := scanner.Text()

		// fmt.Printf("Lines: %s (error)\n", line)
		// line := string(bytes)
		// bytes = bytes[:0]

		// Process line
		// toks := strings.Split(strings.TrimSpace(line), " ")
		// NOTES: helpfully some .obj files are separated by tabs!!! Such a crappy format.
		toks := strings.FieldsFunc(strings.TrimSpace(line), func(r rune) bool {
			switch r {
			case ' ', '\t':
				return true
			}
			return false
		})

		// log.Print(toks)
		if len(toks) < 1 {
			continue
		}

		switch toks[0] {
		case "newmtl":
			name := toks[1]
			//log.Printf("Mtl %v", toks[1])
			mtl = &material.Material{}
			rc.AddMaterial(name, mtl)
			//mtllib.Mtls[toks[1]] = mtl
			//mtl.BSDF.IOR = 1.5
			//mtl.BSDF.Roughness = 0.6

		case "Ke":
			r, err := strconv.ParseFloat(toks[1], 32)
			g, err := strconv.ParseFloat(toks[2], 32)
			b, err := strconv.ParseFloat(toks[3], 32)

			mtl.EDF = &edf.Diffuse{E: [3]float32{float32(r), float32(g), float32(b)}}
			// log.Printf("%v",mesh.Verts)
			// log.Printf("A: %v",math.Vec3{float32(x), float32(y), float32(z)})
			if err != nil {
				return err
			}

		case "Kd":
			r, err := strconv.ParseFloat(toks[1], 32)
			g, err := strconv.ParseFloat(toks[2], 32)
			b, err := strconv.ParseFloat(toks[3], 32)

			if mtl.BSDF[0] == nil {
				mtl.BSDF[0] = &bsdf.Diffuse{Kd: &material.ConstantMap{[3]float32{float32(r), float32(g), float32(b)}}}
			}
			// log.Printf("%v",mesh.Verts)
			// log.Printf("A: %v",math.Vec3{float32(x), float32(y), float32(z)})
			if err != nil {
				return err
			}
		case "map_Kd":

			if len(toks) > 1 {
				mtl.BSDF[0] = &bsdf.Diffuse{Kd: &material.TextureMap{toks[1]}}
			}
			//mtl.BSDF.Diffuse = TextureFile(toks[1])
		}
	}
	return nil
}
