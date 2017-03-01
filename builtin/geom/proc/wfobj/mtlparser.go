// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wfobj

import (
	"bufio"
	"bytes"
	"github.com/jamiec7919/vermeer/builtin/maps"
	"github.com/jamiec7919/vermeer/builtin/shader"
	"github.com/jamiec7919/vermeer/colour"
	"io"
	"math"
	"path/filepath"
	"strconv"
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

func (wfobj *File) parseMTL(r io.Reader, path string) (shaders []*shader.ShaderStd, err error) {
	//var mtlid int
	scanner := bufio.NewScanner(r)
	// bytes := make([]byte, DefaultBufferSize)

	var mtl *shader.ShaderStd

	for scanner.Scan() {
		line := scanner.Text()

		lscan := lineScanner{}
		lscan.init(line)

		cmd := lscan.Token()

		switch cmd {
		case "newmtl":
			name := lscan.Token()

			mtl = &shader.ShaderStd{MtlName: name}
			shaders = append(shaders, mtl)

		case "d":

			d, err := strconv.ParseFloat(lscan.Token(), 32)

			if err != nil {
				return shaders, err
			}

			if d == 0.0 {
				// 'dissolve'
				mtl.DiffuseColour = &maps.Constant{C: colour.RGB{0.5, 0.5, 0.5}, Chan: 0}
				mtl.DiffuseStrength = &maps.Constant{C: colour.RGB{float32(0.5)}, Chan: 0}
			}
		case "Ke":

			r, err := strconv.ParseFloat(lscan.Token(), 32)
			g, err := strconv.ParseFloat(lscan.Token(), 32)
			b, err := strconv.ParseFloat(lscan.Token(), 32)

			if err != nil {
				return shaders, err
			}

			if r == 0.0 && g == 0.0 && b == 0.0 {
				continue
			}

			strength := math.Sqrt(r*r + g*g + b*b)
			r /= strength
			g /= strength
			b /= strength

			mtl.EmissionColour = &maps.Constant{C: colour.RGB{float32(r), float32(g), float32(b)}, Chan: 0}
			mtl.EmissionStrength = &maps.Constant{C: colour.RGB{float32(strength)}, Chan: 0}
			//mtl.params["EmissionColour"] = &rgbparam{[3]float32{float32(r), float32(g), float32(b)}}
			//mtl.params["EmissionStrength"] = &floatparam{float32(strength)}

		case "Kd":

			r, err := strconv.ParseFloat(lscan.Token(), 32)
			g, err := strconv.ParseFloat(lscan.Token(), 32)
			b, err := strconv.ParseFloat(lscan.Token(), 32)

			if r == 0 && g == 0 && b == 0 {

				continue
			}

			//fmt.Printf("diff %v %v %v\n", r, g, b)
			if err != nil {
				return shaders, err
			}

			mtl.DiffuseColour = &maps.Constant{C: colour.RGB{float32(r), float32(g), float32(b)}, Chan: 0}
			mtl.DiffuseStrength = &maps.Constant{C: colour.RGB{float32(0.5)}, Chan: 0}

		case "Ks":

			r, err := strconv.ParseFloat(lscan.Token(), 32)
			g, err := strconv.ParseFloat(lscan.Token(), 32)
			b, err := strconv.ParseFloat(lscan.Token(), 32)

			if err != nil {
				return shaders, err
			}
			//fmt.Printf("%v %v %v\n", r, g, b)

			if r == 0 && g == 0 && b == 0 {
				continue
			}

			mtl.Spec1Colour = &maps.Constant{C: colour.RGB{float32(r), float32(g), float32(b)}, Chan: 0}
			mtl.Spec1Strength = &maps.Constant{C: colour.RGB{float32(0.5)}, Chan: 0}
			mtl.Spec1Roughness = &maps.Constant{C: colour.RGB{float32(0.0)}, Chan: 0}

			//mtl.params["DiffuseColour"] = &rgbparam{[3]float32{float32(r), float32(g), float32(b)}}
			//mtl.params["DiffuseStrength"] = &floatparam{float32(1)}

		case "map_Kd":

			//mtl.params["DiffuseColour"] = &texparam{lscan.Rest()}
			//mtl.params["DiffuseStrength"] = &floatparam{float32(1)}
			fname := lscan.Rest()
			//fmt.Printf("diffmap: %v\n", fname)
			mtl.DiffuseColour = &maps.Texture{Filename: filepath.Join(path, fname), Chan: 0}
			mtl.DiffuseStrength = &maps.Constant{C: colour.RGB{float32(0.5)}, Chan: 0}

		case "map_bump":
		/*
			i := 1
			scale := float32(1.0)
			rest := lscan.Rest()
			if lscan.Token() == "-bm" {
				i++
				scale64, err := strconv.ParseFloat(lscan.Token(), 32)
				scale = float32(scale64)

				if err != nil {
					return shaders, err
				}
				i++
				//mtl.params["BumpMap"] = &texparam{lscan.Rest()}
				//mtl.params["BumpMapScale"] = &floatparam{scale}
			} else {

				//mtl.params["BumpMap"] = &texparam{rest}
				//mtl.params["BumpMapScale"] = &floatparam{scale}

			}
		*/
		default:
			//fmt.Printf("%v\n", mtl)
		}

	}

	return
}
