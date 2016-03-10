package material

import (
	"github.com/jamiec7919/vermeer/internal/core"
	"github.com/jamiec7919/vermeer/material"
	"github.com/jamiec7919/vermeer/material/bsdf"
)

type Node struct {
	material *material.Material
}

func (node *Node) Material() *material.Material            { return node.material }
func (node *Node) Name() string                            { return node.material.Name }
func (node *Node) PreRender(rc *core.RenderContext) error  { return nil }
func (node *Node) PostRender(rc *core.RenderContext) error { return nil }

func makeBSDF2(rc *core.RenderContext, params core.Params) (interface{}, error) {
	bsdf := makeBSDF(params)

	return bsdf, nil
}

func makeBSDF(params core.Params) material.BSDF {
	ty, _ := params.GetString("Type")

	switch ty {
	case "Lambert":
		bsdf := bsdf.Diffuse{}
		Kd := params["Kd"]

		switch t := Kd.(type) {
		case string:
			bsdf.Kd = &material.TextureMap{t}
		case []float64:
			bsdf.Kd = &material.ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		case []int64:
			bsdf.Kd = &material.ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		}
		return &bsdf
	case "OrenNayar":
		bsdf := bsdf.OrenNayar{}

		roughness := params["Roughness"]

		switch t := roughness.(type) {
		case string:
			bsdf.Roughness = &material.TextureMap{t}
		case float64:
			bsdf.Roughness = &material.ConstantMap{[3]float32{float32(t)}}
		case int64:
			bsdf.Roughness = &material.ConstantMap{[3]float32{float32(t)}}
		}

		Kd := params["Kd"]

		switch t := Kd.(type) {
		case string:
			bsdf.Kd = &material.TextureMap{t}
		case []float64:
			bsdf.Kd = &material.ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		case []int64:
			bsdf.Kd = &material.ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		}
		return &bsdf

	case "Specular":
		bsdf := bsdf.Specular{}

		Ks, present := params["Ks"]

		if !present {
			Ks = []float64{0.5, 0.5, 0.5}
		}

		switch t := Ks.(type) {
		case string:
			bsdf.Ks = &material.TextureMap{t}
		case []float64:
			bsdf.Ks = &material.ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		case []int64:
			bsdf.Ks = &material.ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		}
		return &bsdf

	case "GGXSpecular":
		bsdf := bsdf.CookTorranceGGX{}

		roughness := params["Roughness"]

		switch t := roughness.(type) {
		case string:
			bsdf.Roughness = &material.TextureMap{t}
		case float64:
			bsdf.Roughness = &material.ConstantMap{[3]float32{float32(t)}}
		case int64:
			bsdf.Roughness = &material.ConstantMap{[3]float32{float32(t)}}
		}

		ior := params["IOR"]

		switch t := ior.(type) {
		case string:
			bsdf.IOR = &material.TextureMap{t}
		case float64:
			bsdf.IOR = &material.ConstantMap{[3]float32{float32(t)}}
		case int64:
			bsdf.IOR = &material.ConstantMap{[3]float32{float32(t)}}
		}

		Ks, present := params["Ks"]

		if !present {
			Ks = []float64{0.5, 0.5, 0.5}
		}

		switch t := Ks.(type) {
		case string:
			bsdf.Ks = &material.TextureMap{t}
		case []float64:
			bsdf.Ks = &material.ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		case []int64:
			bsdf.Ks = &material.ConstantMap{[3]float32{float32(t[0]), float32(t[1]), float32(t[2])}}
		}
		return &bsdf
	}

	return nil
}

func makeMaterial(rc *core.RenderContext, params core.Params) (interface{}, error) {

	_bsdfs := params["BSDF"]

	bsdfs := _bsdfs.([]interface{}) // Don't like this..

	bsdf := bsdfs[0].(material.BSDF)

	name, _ := params.GetString("Name")
	mtl := &material.Material{Name: name, BSDF: [2]material.BSDF{bsdf}}
	return &Node{mtl}, nil
}

func init() {
	core.RegisterType("BSDF", makeBSDF2)
	core.RegisterType("Material", makeMaterial)
}
