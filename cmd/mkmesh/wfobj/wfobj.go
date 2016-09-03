// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wfobj

import (
	"bufio"
	"fmt"
	m "github.com/jamiec7919/vermeer/math"
	"io"
	"strconv"
	"strings"
)

type facevert struct {
	v, vt, vn int
}

// Mesh represents a single part of a file.
// Each mesh has a single shader
type Mesh struct {
	UV      []m.Vec2
	Verts   []m.Vec3
	Normals []m.Vec3

	UVIdx     []uint32
	VertIdx   []uint32
	NormalIdx []uint32

	PolyCount []uint32

	UVBase     int
	VertBase   int
	NormalBase int

	Shader string
}

func (m *Mesh) WriteNodes(w io.Writer, prefix, name string) error {
	fmt.Fprintf(w, "PolyMesh {\n")
	fmt.Fprintf(w, "	Name \"%v:%v\"\n", prefix, name)
	fmt.Fprintf(w, "	Shader \"%v:%v\"\n", prefix, m.Shader)
	fmt.Fprintf(w, "	RayBias %v\n", 0.1)

	fmt.Fprintf(w, "	Verts 1 %v point", len(m.Verts))

	for i := range m.Verts {
		fmt.Fprintf(w, " %v %v %v", m.Verts[i][0], m.Verts[i][1], m.Verts[i][2])
	}

	fmt.Fprintf(w, "\n	FaceIdx %v int ", len(m.VertIdx))

	for i := range m.VertIdx {
		fmt.Fprintf(w, " %v", m.VertIdx[i])
	}

	fmt.Fprintf(w, "\n	PolyCount %v int ", len(m.PolyCount))

	for i := range m.PolyCount {
		fmt.Fprintf(w, " %v", m.PolyCount[i])
	}

	if len(m.UV) > 0 {
		fmt.Fprintf(w, "\n	UV 1 %v vec2", len(m.UV))

		for i := range m.UV {
			fmt.Fprintf(w, " %v %v", m.UV[i][0], m.UV[i][1])
		}
	}

	if len(m.UVIdx) > 0 {
		fmt.Fprintf(w, "\n	UVIdx %v int", len(m.UVIdx))

		for i := range m.UVIdx {
			fmt.Fprintf(w, " %v", m.UVIdx[i])
		}

	}

	if len(m.Normals) > 0 {
		fmt.Fprintf(w, "\n	Normals 1 %v vec3", len(m.Normals))

		for i := range m.Normals {
			fmt.Fprintf(w, " %v %v %v", m.Normals[i][0], m.Normals[i][1], m.Normals[i][2])
		}

	}

	if len(m.NormalIdx) > 0 {
		fmt.Fprintf(w, "\n	NormalIdx %v int", len(m.NormalIdx))

		for i := range m.NormalIdx {
			fmt.Fprintf(w, " %v", m.NormalIdx[i])
		}

	}

	fmt.Fprintf(w, "\n}\n\n")

	return nil
}

// Load loads the WFObj from the given reader.
func Load(r io.Reader) ([]*Mesh, error) {

	meshes, err := parse(r)
	return meshes, err
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

	return
}

func parse(r io.Reader) ([]*Mesh, error) {
	scanner := bufio.NewScanner(bufio.NewReader(r))

	verts := []m.Vec3{}
	uvs := []m.Vec2{}
	normals := []m.Vec3{}

	meshes := map[string]*Mesh{}

	var mesh *Mesh

	lineno := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineno++

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

		case "usemtl":

			m, present := meshes[toks[1]]

			if !present {
				mesh = &Mesh{Shader: toks[1]}
				meshes[toks[1]] = mesh
			} else {
				mesh = m
			}

		case "f":
			var face []facevert

			for i := 1; i < len(toks); i++ {

				p, n, t := parseFaceField(toks[i])

				//log.Printf("Face %v %v %v %v %v %v", p, t, n, len(reader.v), len(reader.vt), len(reader.vn))
				if p <= 0 {
					p = len(verts) + p + 1
				}

				//				if r.MergeVertPos {
				//					p = int(vertmergeface[p-1])
				//				}

				if t <= 0 {
					t = len(uvs) + t + 1
				}

				//				if r.MergeTexVert {
				//					t = int(texvertmergeface[t-1])
				//				}

				if n <= 0 {
					n = len(normals) + n + 1
				}

				face = append(face, facevert{p - 1, t - 1, n - 1})
			}

			for k := 0; k < len(face)-2; k++ {
				V0 := verts[face[0].v]
				V1 := verts[face[k+1].v]
				V2 := verts[face[k+2].v]

				if V0 == V1 && V0 == V2 {
					fmt.Printf("nil triangle %v %v %v: %v\n", face[0].v, face[k+1].v, face[k+2].v, V0, verts[face[k+1].v], verts[face[k+2].v])
					break
				}

			}

			for _, fvert := range face {
				mesh.VertIdx = append(mesh.VertIdx, uint32(len(mesh.Verts)))
				mesh.Verts = append(mesh.Verts, verts[fvert.v])

				if len(uvs) > 0 {

					if fvert.vt < 0 || fvert.vt >= len(uvs) {
						//fmt.Printf("%v: %v %v\n", lineno, fvert.vt, len(uvs))
						continue
					}

					mesh.UVIdx = append(mesh.UVIdx, uint32(len(mesh.UV)))
					mesh.UV = append(mesh.UV, uvs[fvert.vt])
				}

				if len(normals) > 0 {
					if fvert.vn < 0 || fvert.vn >= len(normals) {
						//fmt.Printf("%v: %v %v\n", lineno, fvert.vn, len(normals))
						continue
					}

					mesh.NormalIdx = append(mesh.NormalIdx, uint32(len(mesh.Normals)))
					mesh.Normals = append(mesh.Normals, normals[fvert.vn])
				}
			}

			mesh.PolyCount = append(mesh.PolyCount, uint32(len(face)))

		case "vt":
			x, err := strconv.ParseFloat(toks[1], 32)

			if err != nil {
				fmt.Printf("Error parsing vertex %v: %v\n", err)

			}

			y, err := strconv.ParseFloat(toks[2], 32)

			if err != nil {
				fmt.Printf("Error parsing vertex %v: %v\n", err)

			}

			uvs = append(uvs, m.Vec2{float32(x), float32(y)})

		case "vn":
			x, err := strconv.ParseFloat(toks[1], 32)

			if err != nil {
				fmt.Printf("Error parsing vertex %v: %v\n", err)

			}

			y, err := strconv.ParseFloat(toks[2], 32)

			if err != nil {
				fmt.Printf("Error parsing vertex %v: %v\n", err)

			}

			z, err := strconv.ParseFloat(toks[3], 32)

			if err != nil {
				fmt.Printf("Error parsing vertex %v: %v\n", err)

			}

			normals = append(normals, m.Vec3{float32(x), float32(y), float32(z)})

		case "v":
			//fmt.Printf("v: %v %v %v %v\n", toks[1], toks[2], toks[3], toks)

			x, err := strconv.ParseFloat(toks[1], 32)

			if err != nil {
				fmt.Printf("Error parsing vertex %v: %v\n", err)

			}

			y, err := strconv.ParseFloat(toks[2], 32)

			if err != nil {
				fmt.Printf("Error parsing vertex %v: %v\n", err)

			}

			z, err := strconv.ParseFloat(toks[3], 32)

			if err != nil {
				fmt.Printf("Error parsing vertex %v: %v\n", err)

			}

			verts = append(verts, m.Vec3{float32(x), float32(y), float32(z)})

		default:
		}
	}

	fmt.Printf("%v %v %v\n", len(verts), len(uvs), len(normals))

	var mshs []*Mesh

	totalfaces := 0

	for _, m := range meshes {
		//		fmt.Printf("%v %v\n", m.Shader, m.PolyCount)
		//		fmt.Printf("%v %v\n", m.Verts, m.VertIdx)
		//		fmt.Printf("%v %v\n", m.Normals, m.NormalIdx)
		//		fmt.Printf("%v %v\n", m.UV, m.UVIdx)
		fmt.Printf("%v %v\n", m.Shader, len(m.PolyCount))
		totalfaces += len(m.PolyCount)
		mshs = append(mshs, m)
	}

	fmt.Printf("TOTAL: %v\n", totalfaces)
	return mshs, nil
}
