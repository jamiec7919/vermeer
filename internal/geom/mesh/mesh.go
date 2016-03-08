// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mesh

import (
	"github.com/jamiec7919/vermeer/internal/core"
	m "github.com/jamiec7919/vermeer/math"
	"github.com/jamiec7919/vermeer/qbvh"
	"log"
)

// WARNING: Do not modify this without careful consideration! Designed to fit neatly in cache.
// 16 float32, 64bytes
type FaceGeom struct {
	V     [3]m.Vec3
	N     m.Vec3   // Not actually used in intersection code at moment..
	Vi    [3]int32 // reference the tex coords etc.
	MtlId int32    // Material Id
}

func (f *FaceGeom) setup() {
	f.N = m.Vec3Normalize(m.Vec3Cross(m.Vec3Sub(f.V[1], f.V[0]), m.Vec3Sub(f.V[2], f.V[0])))

	for i := range f.N {
		if f.N[i] == 0.0 {
			f.N[i] = 0
		}
	}
}

type Loader interface {
	Load() (*Mesh, error)
}

type Mesh struct {
	Faces     []FaceGeom
	Vn        []m.Vec3
	Vuv       [][]m.Vec2
	nodes     []qbvh.Node
	faceindex []int32 // Face indexes - used only with acceleration leaf structure, may contain duplicates
	bounds    m.BoundingBox
}

type StaticMesh struct {
	NodeName string
	Mesh     *Mesh
}

func (mesh *StaticMesh) Name() string { return mesh.NodeName }
func (mesh *StaticMesh) PreRender(rc *core.RenderContext) error {
	return mesh.Mesh.initAccel()
}
func (mesh *StaticMesh) PostRender(rc *core.RenderContext) error { return nil }

func (mesh *StaticMesh) TraceRay(ray *core.RayData) {
	mesh.Mesh.traceRayAccel(ray)
}

func (mesh *StaticMesh) VisRay(ray *core.RayData) {
	mesh.Mesh.visRayAccel(ray)
}

type MeshFile struct {
	NodeName string
	Filename string
	Loader   Loader
	mesh     *Mesh
}

func (mesh *MeshFile) Name() string { return mesh.NodeName }
func (mesh *MeshFile) PreRender(rc *core.RenderContext) error {
	msh, err := mesh.Loader.Load()

	if err != nil {
		return err
	}

	if mesh.NodeName == "mesh02" {
		trn := m.Matrix4Translate(0.4, -0.4, -0.4)
		rot := m.Matrix4Rotate(m.Pi/2, 0, 1, 0)
		msh.Transform(m.Matrix4Mul(trn, rot))
	}
	mesh.mesh = msh
	return mesh.mesh.initAccel()
}
func (mesh *MeshFile) PostRender(rc *core.RenderContext) error { return nil }

func (mesh *MeshFile) TraceRay(ray *core.RayData) {
	mesh.mesh.traceRayAccel(ray)
}

func (mesh *MeshFile) VisRay(ray *core.RayData) {
	mesh.mesh.visRayAccel(ray)
}

func (mesh *Mesh) Transform(trn m.Matrix4) {
	for i := range mesh.Faces {
		mesh.Faces[i].V[0] = m.Matrix4MulPoint(trn, mesh.Faces[i].V[0])
		mesh.Faces[i].V[1] = m.Matrix4MulPoint(trn, mesh.Faces[i].V[1])
		mesh.Faces[i].V[2] = m.Matrix4MulPoint(trn, mesh.Faces[i].V[2])
	}
}
func (mesh *Mesh) WorldBounds(m m.Matrix4) (out m.BoundingBox) {
	return
}

func calcBox(verts []m.Vec3) (bounds m.BoundingBox) {
	bounds.Reset()

	for i := range verts {
		bounds.GrowVec3(verts[i])
	}
	return
}

/*
Change this for early split clipping, check each triangle BB for surface area, if greater than a maximum then
split along longest axis, and add two new bounding boxes+indices to original triangle.  Need to reintroduce
indices or copy triangles. Maybe use median BB surface area for mesh as maximum.
*/
func trisplit(verts []m.Vec3, idx int32, indexes *[]int32, boxes *[]m.BoundingBox, centroids *[]m.Vec3) {
	var stack [][]m.Vec3
	stack = append(stack, verts)
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		var box m.BoundingBox
		box.Reset()

		if len(top) < 1 {
			panic("<1")
		}
		for i := range top {
			box.GrowVec3(top[i])
		}

		if box.SurfaceArea() > 50000000 {
			axis := box.MaxDim()

			d := box.Centroid()[axis]

			//log.Printf("SA: %v %v %v", box.SurfaceArea(), axis, d)
			primng := qbvh.Clipleft(d, axis, top)
			//log.Printf("%v", primng)
			primpv := qbvh.Clipright(d, axis, top)
			stack = append(stack, primng)
			//log.Printf("%v", primpv)
			stack = append(stack, primpv)
		} else {
			*indexes = append(*indexes, idx)
			*boxes = append(*boxes, box)

			var centroid m.Vec3
			for i := range top {
				centroid = m.Vec3Add(centroid, top[i])
			}

			centroid = m.Vec3Scale(1.0/float32(len(top)), centroid)
			*centroids = append(*centroids, centroid)

			//log.Printf("Add: %v %v %v", idx, box, centroid)
		}
	}
}

func (mesh *Mesh) initAccel() error {
	for face := range mesh.Faces {
		mesh.Faces[face].setup()
		//log.Printf("%v %v ", m.Faces[face].N, m.Faces[face].MtlId)
	}

	boxes := make([]m.BoundingBox, 0, len(mesh.Faces))
	centroids := make([]m.Vec3, 0, len(mesh.Faces))
	indxs := make([]int32, 0, len(mesh.Faces))

	for i := range mesh.Faces {
		verts := make([]m.Vec3, 3)
		verts[0] = mesh.Faces[i].V[0]
		verts[1] = mesh.Faces[i].V[1]
		verts[2] = mesh.Faces[i].V[2]
		trisplit(verts, int32(i), &indxs, &boxes, &centroids)
		/*box := m.BoundingBox{}
		box.Reset()
		box.Grow(m.Faces[i].V[0][0], m.Faces[i].V[0][1], m.Faces[i].V[0][2])
		box.Grow(m.Faces[i].V[1][0], m.Faces[i].V[1][1], m.Faces[i].V[1][2])
		box.Grow(m.Faces[i].V[2][0], m.Faces[i].V[2][1], m.Faces[i].V[2][2])
		centroid := m.Vec3{}
		centroid[0] = (m.Faces[i].V[0][0] + m.Faces[i].V[1][0] + m.Faces[i].V[2][0]) / 3
		centroid[1] = (m.Faces[i].V[0][1] + m.Faces[i].V[1][1] + m.Faces[i].V[2][1]) / 3
		centroid[2] = (m.Faces[i].V[0][2] + m.Faces[i].V[1][2] + m.Faces[i].V[2][2]) / 3
		indxs = append(indxs, int32(i))
		boxes = append(boxes, box)
		centroids = append(centroids, centroid)
		*/
	}

	//for i := range boxes {
	//log.Printf("%v", boxes[i].SurfaceArea())
	//	}
	//panic("stop")
	nodes, bounds := qbvh.BuildAccel(boxes, centroids, indxs, 16)

	mesh.nodes = nodes
	mesh.bounds = bounds

	//newfaces := make([]FaceGeom, len(m.Faces))

	// rearrange faces to match index structure of nodes
	//for i := range indxs {
	//		newfaces[i] = m.Faces[indxs[i]]
	//	}

	totalleafs := 0

	nodef := func(i int, bounds m.BoundingBox) { /*log.Printf("Node %v", i) */ }
	leaf := func(bounds m.BoundingBox, base, count int, empty bool) {
		totalleafs++
		if !empty {
			//			log.Printf("Leaf %v %v %v %v", bounds, base, count, bounds.SurfaceArea())
		} else {
			//			log.Printf("Leaf empty")
		}
	}
	qbvh.Walk(nodes, 0, nodef, leaf)
	log.Printf("Total leafs: %v", totalleafs)
	//	m.Faces = newfaces
	mesh.faceindex = indxs
	return nil
}

var loaders = map[string]func(rc *core.RenderContext, filename string) (Loader, error){}

func RegisterLoader(name string, open func(rc *core.RenderContext, filename string) (Loader, error)) {
	loaders[name] = open
}

func create(rc *core.RenderContext, params core.Params) (interface{}, error) {
	if filename, present := params["Filename"]; present {
		for _, open := range loaders {
			loader, err := open(rc, filename.(string))

			if err == nil {
				name := params["Name"]

				if name == nil {
					name = "<mesh>"
				}

				mesh := MeshFile{NodeName: name.(string), Filename: filename.(string), Loader: loader}

				//rc.AddNode(&mesh)
				return &mesh, nil
			}
		}
	}
	return nil, nil
}

func init() {
	core.RegisterType("MeshFile", create)
}