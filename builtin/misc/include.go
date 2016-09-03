package misc

import (
	"github.com/jamiec7919/vermeer/core"
	"github.com/jamiec7919/vermeer/nodes"
)

// Include represents a circular disk light node.
type Include struct {
	NodeDef  core.NodeDef
	NodeName string `node:"Name"`
	Filename string
}

var _ core.Node = (*Include)(nil)

// Name implements core.Node.
func (d *Include) Name() string { return d.NodeName }

// Def implements core.Node.
func (d *Include) Def() core.NodeDef { return d.NodeDef }

// PreRender implelments core.Node.
func (d *Include) PreRender() error {
	return nodes.Parse(d.Filename)
}

// PostRender implelments core.Node.
func (d *Include) PostRender() error { return nil }

func init() {
	nodes.Register("Include", func() (core.Node, error) {

		return &Include{}, nil

	})
}
