// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wfobj

import (
	"bufio"
	//"fmt"
	"github.com/jamiec7919/vermeer/internal/core"
	"github.com/jamiec7919/vermeer/internal/geom/mesh"
	"github.com/jamiec7919/vermeer/material"
	m "github.com/jamiec7919/vermeer/math"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

type facevert struct {
	p, t, n int
}

type texvert struct {
	t, n int
}

type reader struct {
	v     []m.Vec3
	vt    []m.Vec2
	vn    []m.Vec3
	vn_i  map[texvert]int32
	Faces []mesh.FaceGeom

	mvt []m.Vec2
	mvn []m.Vec3

	areas []float64

	rc       *core.RenderContext
	filename string
}

func (r *reader) init(rc *core.RenderContext, filename string) {
	r.rc = rc
	r.filename = filename
	r.vn_i = make(map[texvert]int32)
}

//TODO: this is a bit of a mess
func (r *reader) triangulateFace(face []facevert, mtlid material.Id) {
	i := 1

	for k := 0; k < len(face)-2; k++ {
		i2 := i + 1

		f := mesh.FaceGeom{}
		//log.Printf("k %v %v %v %v %v %v %v %v %v %v %v", len(r.Faces), k, face[0].p, face[i].p, face[i2].p, len(r.v), len(r.vt), len(r.vn), face[0].t, face[i].t, face[i2].t)
		f.V[0] = r.v[face[0].p]
		f.V[1] = r.v[face[i].p]
		f.V[2] = r.v[face[i2].p]

		if tv, ok := r.vn_i[texvert{face[0].t, face[0].n}]; ok {
			// We have a match for this tv, use that index
			f.Vi[0] = int32(tv)
		} else {
			// alloc a new one
			f.Vi[0] = int32(len(r.mvt))
			texvert := texvert{face[0].t, face[0].n}
			r.vn_i[texvert] = f.Vi[0]

			if r.vt != nil {
				if face[0].t < len(r.vt) {
					r.mvt = append(r.mvt, r.vt[face[0].t])
				} else {
					r.mvt = append(r.mvt, m.Vec2{})
				}
			} else {
				r.mvt = append(r.mvt, m.Vec2{})

			}
			if r.vn != nil {
				r.mvn = append(r.mvn, r.vn[face[0].n])
			}
		}

		if tv, ok := r.vn_i[texvert{face[i].t, face[i].n}]; ok {
			// We have a match for this tv, use that index
			f.Vi[1] = int32(tv)
		} else {
			// alloc a new one
			f.Vi[1] = int32(len(r.mvt))
			texvert := texvert{face[i].t, face[i].n}
			r.vn_i[texvert] = f.Vi[1]

			if r.vt != nil {
				if face[i].t < len(r.vt) {
					r.mvt = append(r.mvt, r.vt[face[i].t])
				} else {
					r.mvt = append(r.mvt, m.Vec2{})
				}
			} else {
				r.mvt = append(r.mvt, m.Vec2{})

			}
			if r.vn != nil {
				r.mvn = append(r.mvn, r.vn[face[i].n])
			}
		}

		if tv, ok := r.vn_i[texvert{face[i2].t, face[i2].n}]; ok {
			// We have a match for this tv, use that index
			f.Vi[2] = int32(tv)
		} else {
			// alloc a new one
			f.Vi[2] = int32(len(r.mvt))
			texvert := texvert{face[i2].t, face[i2].n}
			r.vn_i[texvert] = f.Vi[2]
			if r.vt != nil {
				if face[i2].t < len(r.vt) {
					r.mvt = append(r.mvt, r.vt[face[i2].t])
				} else {
					r.mvt = append(r.mvt, m.Vec2{})
				}
			} else {
				r.mvt = append(r.mvt, m.Vec2{})

			}
			if r.vn != nil {
				r.mvn = append(r.mvn, r.vn[face[i2].n])
			}

		}
		i++

		f.MtlId = int32(mtlid)
		//f.setup()

		area := float64(m.Vec3Length(m.Vec3Cross(m.Vec3Sub(f.V[2], f.V[0]), m.Vec3Sub(f.V[1], f.V[0]))))
		if area > 100000 {
			//log.Printf("Warning: over large triangle %v", area)
			//continue
		}
		r.areas = append(r.areas, area)
		r.Faces = append(r.Faces, f)
	}
}

func (r *reader) printAreas() {
	sort.Float64s(r.areas)

	for _, area := range r.areas {
		log.Printf("%v", area)
	}

	log.Printf("Median: %v", r.areas[len(r.areas)/2])

}

func Open(rc *core.RenderContext, filename string) (mesh.Loader, error) {
	fin, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	fin.Close()
	reader := reader{}
	reader.init(rc, filename)

	return &reader, nil
}

func parseFaceField(field string) (p, n, t int) {
	fields := strings.Split(field, "/")

	p, _ = strconv.Atoi(fields[0])

	if len(fields) > 1 {
		if len(fields[1]) > 0 {
			t, _ = strconv.Atoi(fields[1])
		}
	}

	if len(fields) > 2 {
		if len(fields[2]) > 0 {
			n, _ = strconv.Atoi(fields[2])
		}
	}
	/*
		fmt.Sscanf(field, "%d/%d/%d", &p, &t, &n)

		if p == 0 {
			fmt.Sscanf(field, "%d", &p)
			//fmt.Printf("p==0\n")
		}
		if t == 0 {
			fmt.Sscanf(field, "%d//%d", &p, &n)
			//fmt.Printf("t==0 %v %v\n", p, n)
		}
		if n == 0 {

		}
	*/
	return

}
func (r *reader) Load() (out *mesh.Mesh, err error) {
	fin, err := os.Open(r.filename)
	if err != nil {
		return nil, err
	}

	defer fin.Close()

	scanner := bufio.NewScanner(bufio.NewReader(fin))
	//scanner := bufio.NewScanner(fin)

	// We will use the vt indices as given and remap the vn's.
	//vn_i := map[int]int{}

	var mtlid material.Id = material.ID_NONE

	for scanner.Scan() {
		line := scanner.Text()

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
		case "mtllib":
			err := ParseMtlLib(r.rc, toks[1])

			if err != nil {
				return nil, err
			}
		case "usemtl":
			//fmt.Printf("Mtl %v\n", toks[1])
			mtlid = r.rc.GetMaterialId(toks[1])

			//TODO: Should check if this material is emissive and if so then create a new mesh lightsource.

		case "f":
			if len(toks) > 4 {
				//log.Printf("%v", len(toks))
			}

			var face []facevert

			for i := 1; i < len(toks); i++ {

				p, n, t := parseFaceField(toks[i])

				//log.Printf("Face %v %v %v %v %v %v", p, t, n, len(reader.v), len(reader.vt), len(reader.vn))
				if p > 0 {
					//		p -= 1
				} else {
					p = len(r.v) + p + 1
				}

				if t > 0 {
					//		t -= 1
				} else {
					t = len(r.vt) + t + 1

				}
				if n > 0 {
					//		n -= 1
				} else {
					n = len(r.vn) + n + 1

				}

				face = append(face, facevert{p - 1, t - 1, n - 1})
				/*
					toks2 := strings.Split(toks[1+i], "/")
					// log.Printf("toks2: %v",toks2)
					p, err := strconv.ParseInt(toks2[0], 10, 32)

					t := int64(0)

					if len(toks2) > 1 {
						t, err = strconv.ParseInt(toks2[1], 10, 32)
					}

					//					t, err := strconv.ParseInt(toks2[1], 10, 32)
					n := int64(0)

					if len(toks2) > 2 {
						n, err = strconv.ParseInt(toks2[2], 10, 32)
					}
				*/
			}

			if mtlid == material.ID_NONE {
				// assign 'default' material
				mtlid = r.rc.GetMaterialId("mtl_default")
				//log.Printf("setting default mtl %v", mtlid)
			}
			r.triangulateFace(face, mtlid)
		case "vt":
			// log.Printf("%v %v %v %v",toks[1],toks[2],toks[3],toks)
			x, err := strconv.ParseFloat(toks[1], 32)
			y, err := strconv.ParseFloat(toks[2], 32)

			r.vt = append(r.vt, m.Vec2{float32(x), float32(y)})
			// log.Printf("%v",mesh.Verts)
			// log.Printf("A: %v",math.Vec3{float32(x), float32(y), float32(z)})
			if err != nil {
				return nil, err
			}
		case "vn":
			// log.Printf("%v %v %v %v",toks[1],toks[2],toks[3],toks)
			x, err := strconv.ParseFloat(toks[1], 32)
			y, err := strconv.ParseFloat(toks[2], 32)
			z, err := strconv.ParseFloat(toks[3], 32)
			//var x, y, z float32
			//fmt.Sscanf(toks[1], "%f", &x)
			//fmt.Sscanf(toks[2], "%f", &y)
			//fmt.Sscanf(toks[3], "%f", &z)

			r.vn = append(r.vn, m.Vec3{float32(x), float32(y), float32(z)})
			// log.Printf("%v",mesh.Verts)
			// log.Printf("A: %v",math.Vec3{float32(x), float32(y), float32(z)})
			if err != nil {
				return nil, err
			}
		case "v":
			// log.Printf("%v %v %v %v",toks[1],toks[2],toks[3],toks)
			//var x, y, z float32
			x, err := strconv.ParseFloat(toks[1], 32)
			y, err := strconv.ParseFloat(toks[2], 32)
			z, err := strconv.ParseFloat(toks[3], 32)
			//fmt.Sscanf(toks[1], "%f", &x)
			//fmt.Sscanf(toks[2], "%f", &y)
			//fmt.Sscanf(toks[3], "%f", &z)
			r.v = append(r.v, m.Vec3{float32(x), float32(y), float32(z)})
			// log.Printf("%v",mesh.Verts)
			// log.Printf("A: %v",math.Vec3{float32(x), float32(y), float32(z)})
			if err != nil {
				return nil, err
			}

		default:
		}
	}
	//reader.printAreas()
	out = &mesh.Mesh{}

	out.Faces = r.Faces
	out.Vn = r.mvn

	if r.mvt != nil {
		out.Vuv = [][]m.Vec2{r.mvt}
	}

	return
}

func init() {
	mesh.RegisterLoader("wfobj", Open)
}
