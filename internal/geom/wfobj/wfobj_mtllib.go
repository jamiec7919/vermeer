// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wfobj

import (
	"bufio"
	"bytes"
	"github.com/jamiec7919/vermeer/colour"
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/material"
	"os"
	"strconv"
	//"strings"
	//"log"
	"unicode/utf8"
)

type lineScanner struct {
	line []byte
	pos  int
}

func (l *lineScanner) next() rune {
	if len(l.line) == 0 {
		return 0
	}

	c, size := utf8.DecodeRune(l.line)
	l.line = l.line[size:]
	if c == utf8.RuneError && size == 1 {
		//log.Print("invalid utf8")
		return l.next()
	}
	return c
}
func (l *lineScanner) init(line string) {
	l.line = []byte(line)
}

func (l *lineScanner) Rest() string {
	//log.Printf("rest: %v", string(l.line))
	return string(l.line)
}

func (l *lineScanner) Token() string {
	var buf bytes.Buffer

	// skip whitespace
L:
	for {
		switch r := l.next(); r {
		case 0, '#':
			return ""

		case ' ', '\t':
			// do nothing, skip
		default:
			buf.WriteRune(r)
			break L
		}
	}

L2:
	for {
		switch r := l.next(); r {
		case 0, ' ', '\t':
			break L2
		default:
			buf.WriteRune(r)
		}
	}
	//log.Printf("tok: %v", buf.String())
	return buf.String()
}

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

		lscan := lineScanner{}
		lscan.init(line)
		// fmt.Printf("Lines: %s (error)\n", line)
		// line := string(bytes)
		// bytes = bytes[:0]

		// Process line
		// toks := strings.Split(strings.TrimSpace(line), " ")
		// NOTES: helpfully some .obj files are separated by tabs!!! Such a crappy format.

		cmd := lscan.Token()

		switch cmd {
		case "newmtl":
			name := lscan.Token()

			//log.Printf("Mtl %v", toks[1])
			mtl = &material.Material{MtlName: name}
			rc.AddNode(mtl)
			//mtllib.Mtls[toks[1]] = mtl
			//mtl.BSDF.IOR = 1.5
			//mtl.BSDF.Roughness = 0.6

		case "Ke":
			r, err := strconv.ParseFloat(lscan.Token(), 32)
			g, err := strconv.ParseFloat(lscan.Token(), 32)
			b, err := strconv.ParseFloat(lscan.Token(), 32)

			// Stupidly some .mtls have Ke but set to 0
			if r > 0.0 || g > 0.0 || b > 0.0 {

				mtl.E = &core.ConstantMap{[3]float32{float32(r), float32(g), float32(b)}}
			}
			// log.Printf("%v",mesh.Verts)
			// log.Printf("A: %v",math.Vec3{float32(x), float32(y), float32(z)})
			if err != nil {
				return err
			}

		case "Kd":
			r, err := strconv.ParseFloat(lscan.Token(), 32)
			g, err := strconv.ParseFloat(lscan.Token(), 32)
			b, err := strconv.ParseFloat(lscan.Token(), 32)

			mtl.Diffuse = "Lambert"
			mtl.Kd = &core.ConstantMap{[3]float32{float32(r), float32(g), float32(b)}}
			// log.Printf("%v",mesh.Verts)
			// log.Printf("A: %v",math.Vec3{float32(x), float32(y), float32(z)})
			if err != nil {
				return err
			}
		case "Ks":
			r, err := strconv.ParseFloat(lscan.Token(), 32)
			g, err := strconv.ParseFloat(lscan.Token(), 32)
			b, err := strconv.ParseFloat(lscan.Token(), 32)

			mtl.Specular = "Specular"

			rgb := colour.RGB{float32(r), float32(g), float32(b)}
			rgb.Normalize()
			mtl.Ks = &core.ConstantMap{[3]float32(rgb)}
			// log.Printf("%v",mesh.Verts)
			// log.Printf("A: %v",math.Vec3{float32(x), float32(y), float32(z)})
			if err != nil {
				return err
			}

		case "map_Kd":

			mtl.Diffuse = "Lambert"
			mtl.Kd = &core.TextureMap{lscan.Rest()}
			//mtl.BSDF.Diffuse = TextureFile(toks[1])
		case "map_bump":
			i := 1
			scale := float32(1.0)
			rest := lscan.Rest()
			if lscan.Token() == "-bm" {
				i++
				scale64, err := strconv.ParseFloat(lscan.Token(), 32)
				scale = float32(scale64)

				if err != nil {
					return err
				}
				i++

				filename := lscan.Rest()

				if filename != "" {
					mtl.BumpMap = &core.TextureMap{filename}
					mtl.BumpMapScale = scale
				}
			} else {
				if rest != "" {
					mtl.BumpMap = &core.TextureMap{rest}
					mtl.BumpMapScale = scale
				}
			}

			//if i < len(toks) {
			//}
		}
	}
	return nil
}
