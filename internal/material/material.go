package material

import (
	"github.com/jamiec7919/vermeer/internal/core"
	"github.com/jamiec7919/vermeer/internal/nodeparser"
	"github.com/jamiec7919/vermeer/material"
	"github.com/jamiec7919/vermeer/material/bsdf"
)

func makeBSDF(params core.Params) material.BSDF {
	ty, _ := params.GetString("Type")

	switch ty {
	case "Diffuse":
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
	}

	return nil
}

func makeMaterial(rc *core.RenderContext, params core.Params) error {

	_bsdfs := params["BSDF"]

	bsdfs := _bsdfs.([]nodeparser.Params) // Don't like this..

	bsdf := makeBSDF(core.Params(bsdfs[0]))

	name, _ := params.GetString("Name")

	rc.AddMaterial(name, &material.Material{BSDF: [2]material.BSDF{bsdf}})
	return nil
}

func init() {
	core.RegisterNodeType("Material", makeMaterial)
}
