package core

import (
	"github.com/jamiec7919/vermeer/material"
)

type Material interface {
	Name() string
	Material() *material.Material
}
