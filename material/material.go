package material

import (
	"github.com/jamiec7919/vermeer/material/texture"
)

type Id int32

const ID_NONE Id = -1

type Material struct {
	Name  string
	Sides int
	BSDF  [2]BSDF // bsdf for sides
	//Medium [2]Medium  // medium material
	EDF EDF // Emission distribution for light materials (nil is not a light)
}

type ConstantMap struct {
	C [3]float32
}

func (c *ConstantMap) SampleRGB(s, t, ds, dt float32) (out [3]float32) {
	out[0] = c.C[0]
	out[1] = c.C[1]
	out[2] = c.C[2]
	return
}

func (c *ConstantMap) SampleScalar(s, t, ds, dt float32) (out float32) {
	out = c.C[0]
	return
}

type TextureMap struct {
	Filename string
}

func (c *TextureMap) SampleRGB(s, t, ds, dt float32) (out [3]float32) {
	return texture.SampleRGB(c.Filename, s, t, ds, dt)
}

func (c *TextureMap) SampleScalar(s, t, ds, dt float32) (out float32) {
	q := texture.SampleRGB(c.Filename, s, t, ds, dt)
	return q[0]

}
