// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wfobj

import (
	"bufio"
	"fmt"
	"github.com/jamiec7919/vermeer/builtin/geom/polymesh"
	"github.com/jamiec7919/vermeer/builtin/shader"
	m "github.com/jamiec7919/vermeer/math"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

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

func parseFace(lscan *lineScanner, mesh *polymesh.PolyMesh, mtl byte) error {

	verts := 0

	for {

		field := lscan.Token()

		if field == "" {
			break
		}

		p, n, t := parseFaceField(field)

		//log.Printf("%v %v %v", p, n, t)
		switch {
		case p < 0:
			p = len(mesh.Verts.Elems) + p + 1
		case p == 0:
			p = 1
		}

		//				if r.MergeVertPos {
		//					p = int(vertmergeface[p-1])
		//				}
		switch {
		case t < 0:
			t = len(mesh.UV.Elems) + t + 1
		case t == 0:
			t = 1
		}

		//				if r.MergeTexVert {
		//					t = int(texvertmergeface[t-1])
		//				}
		switch {
		case n < 0:
			n = len(mesh.Normals.Elems) + n + 1

		case n == 0:
			n = 1
		}

		mesh.FaceIdx = append(mesh.FaceIdx, int32(p-1))

		// Annoyingly some .obj files don't have normals until late in the file but the indexes must
		// exist for all triangles!
		//if mesh.Normals.Elems != nil {
		mesh.NormalIdx = append(mesh.NormalIdx, int32(n-1))
		//}
		// Annoyingly some .obj files don't have tex coords until late in the file but the indexes must
		// exist for all triangles!
		//if mesh.UV.Elems != nil {
		mesh.UVIdx = append(mesh.UVIdx, int32(t-1))
		//}
		verts++
	}

	mesh.PolyCount = append(mesh.PolyCount, int32(verts))
	mesh.ShaderIdx = append(mesh.ShaderIdx, int32(mtl))
	return nil
}

func (wfobj *File) parse(r io.Reader) (mesh *polymesh.PolyMesh, shaders []*shader.ShaderStd, err error) {

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024*64), 64)
	lineno := 0

	mesh = &polymesh.PolyMesh{}

	mesh.Verts.MotionKeys = 1
	mesh.Normals.MotionKeys = 1
	mesh.UV.MotionKeys = 1

	mtllibs := map[string]bool{}

	mtl := byte(0)

	for scanner.Scan() {
		line := scanner.Text()
		lineno++

		lscan := lineScanner{}
		lscan.init(line)

		cmd := lscan.Token()

		switch cmd {
		case "mtllib":
			filename := lscan.Rest()

			path := filepath.Join(filepath.Dir(wfobj.Filename), filename)

			if _, present := mtllibs[path]; present {
				continue
			}

			f, err := os.Open(path)

			if err != nil {
				log.Printf("wfobj.parse: Error opening \"%v\": %v", path, err)
				continue
			}

			s, err := wfobj.parseMTL(f, filepath.Dir(wfobj.Filename))

			f.Close()

			if err != nil {
				log.Printf("wfobj.parse: Error parsing \"%v\": %v", path, err)
				continue
			}

			shaders = append(shaders, s...)

			if len(shaders) > 255 {
				return nil, nil, fmt.Errorf("wfobj.parse: too many shaders")
			}

			mtllibs[path] = true

		case "usemtl":

			name := lscan.Rest()

			for i, shader := range shaders {
				if shader.MtlName == name {
					mtl = byte(i)
				}
			}
		case "f":
			if err := parseFace(&lscan, mesh, mtl); err != nil {
				log.Printf("Error parsing face: %v", err)
			}

		case "v":
			x, err := strconv.ParseFloat(lscan.Token(), 32)

			if err != nil {
				log.Printf("Error parsing vertex %v: %v", len(mesh.Verts.Elems)+1, err)

			}

			y, err := strconv.ParseFloat(lscan.Token(), 32)

			if err != nil {
				log.Printf("Error parsing vertex %v: %v", len(mesh.Verts.Elems)+1, err)

			}

			z, err := strconv.ParseFloat(lscan.Token(), 32)

			if err != nil {
				log.Printf("Error parsing vertex %v: %v", len(mesh.Verts.Elems)+1, err)

			}

			mesh.Verts.Elems = append(mesh.Verts.Elems, m.Vec3{float32(x), float32(y), float32(z)})
			mesh.Verts.ElemsPerKey++

		case "vn":
			x, err := strconv.ParseFloat(lscan.Token(), 32)

			if err != nil {
				fmt.Printf("Error parsing vertex vn %v: %v\n", len(mesh.Normals.Elems)+1, err)

			}

			y, err := strconv.ParseFloat(lscan.Token(), 32)

			if err != nil {
				fmt.Printf("Error parsing vertex vn %v: %v\n", len(mesh.Normals.Elems)+1, err)

			}

			z, err := strconv.ParseFloat(lscan.Token(), 32)

			if err != nil {
				fmt.Printf("Error parsing vertex vn %v: %v\n", len(mesh.Normals.Elems)+1, err)

			}

			mesh.Normals.Elems = append(mesh.Normals.Elems, m.Vec3{float32(x), float32(y), float32(z)})
			mesh.Normals.ElemsPerKey++

		case "vt":
			x, err := strconv.ParseFloat(lscan.Token(), 32)

			if err != nil {
				fmt.Printf("Error parsing vertex vt %v: %v\n", len(mesh.UV.Elems)+1, err)

			}

			y, err := strconv.ParseFloat(lscan.Token(), 32)

			if err != nil {
				fmt.Printf("Error parsing vertex vt %v: %v\n", len(mesh.UV.Elems)+1, err)

			}

			mesh.UV.Elems = append(mesh.UV.Elems, m.Vec2{float32(x), float32(y)})
			mesh.UV.ElemsPerKey++

		default:
			//Comment or blank line
		}
	}

	err = nil

	// If there are no Normals then delete the indexes
	if mesh.Normals.Elems == nil {
		mesh.NormalIdx = nil
	}

	// If there are no UVs then delete the indexes
	if mesh.UV.Elems == nil {
		mesh.UVIdx = nil
	}

	//log.Printf("%v %v", len(mesh.FaceIdx), mesh.FaceIdx)
	return
}
