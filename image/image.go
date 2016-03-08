// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	Package image implements a Go-centric version of the Open Image IO library.
*/
package image

import (
	"errors"
)

var (
	ErrNoWriter = errors.New("No writer")
)

type BaseType int
type Aggregate int

const (
	UINT8 BaseType = iota
	FLOAT
)

const (
	SCALAR Aggregate = iota
)

const (
	AutoStride = -1
)

// Defaults to UINT8, SCALAR
type TypeDesc struct {
	BaseType  BaseType
	Aggregate Aggregate
}

type Spec struct {
	Width, Height, Depth             int
	X, Y, Z                          int
	FullWidth, FullHeight, FullDepth int
	FullX, FullY, FullZ              int
	TileWidth, TileHeight, TileDepth int
	NChannels                        int
	Format                           []TypeDesc
	ChannelNames                     []string
	AlphaChannel                     int
	ZChannel                         int
	Deep                             int
	ExtraAttribs                     map[string]interface{}
}

type Reader interface {
	Spec() (Spec, error)
	Close()

	ReadImage(ty TypeDesc, buf interface{}) error
	ReadScanline(y, z int, ty TypeDesc, buf interface{}) error
	ReadTile(x, y, z int, ty TypeDesc, buf interface{}) error

	Supports(tag string) bool
}

type Writer interface {
	Open(string, *Spec) error
	OpenMode(string, *Spec, string) error

	Close()

	WriteImage(ty TypeDesc, buf interface{}) error
	WriteImageStride(ty TypeDesc, buf interface{}, xstride, ystride, zstride int) error
	WriteScanline(y, z int, ty TypeDesc, buf interface{}) error
	WriteScanlineStride(y, z int, ty TypeDesc, buf interface{}, xstride, ystride, zstride int) error
	WriteTile(x, y, z int, ty TypeDesc, buf interface{}) error
	WriteTileStride(x, y, z int, ty TypeDesc, buf interface{}, xstride, ystride, zstride int) error

	Supports(tag string) bool

	// Copy(Reader) error
}

var writers []func(filename string) (Writer, error)
var readers []func(filename string) (Reader, error)

func NewWriter(filename string) (Writer, error) {
	for _, w := range writers {
		writer, err := w(filename)

		if writer != nil && err == nil {
			return writer, err
		}
	}

	return nil, errors.New("No Writer")
}
func Open(filename string) (Reader, error) {
	for _, reader := range readers {

		r, err := reader(filename)

		if err == nil && r != nil {
			return r, err
		}

	}

	return nil, errors.New("image.Open: no reader found for " + filename)
}

func RegisterReader(open func(filename string) (Reader, error)) {
	readers = append(readers, open)
}

func RegisterWriter(test func(filename string) (Writer, error)) {
	writers = append(writers, test)
}
