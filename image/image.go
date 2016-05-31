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
	// ErrNoWriter is returned if no writer can be found for image file.
	ErrNoWriter = errors.New("No writer")
)

// BaseType represents the base type.
type BaseType int

// Aggregate represents the kind of aggregate (e.g. scalar, array etc.).
type Aggregate int

// Enum for BaseTypes.
const (
	UINT8 BaseType = iota
	FLOAT
)

// Enum for Aggregate.
const (
	SCALAR Aggregate = iota
)

const (
	// AutoStride is used in stride parameters to signify default stride.
	AutoStride = -1
)

// TypeDesc describes a type.
// Defaults to UINT8, SCALAR
type TypeDesc struct {
	BaseType  BaseType
	Aggregate Aggregate
}

// Spec describes an image.
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

// Reader is implemented by image readers.
type Reader interface {
	Spec() (Spec, error)
	Close()

	ReadImage(ty TypeDesc, buf interface{}) error
	ReadScanline(y, z int, ty TypeDesc, buf interface{}) error
	ReadTile(x, y, z int, ty TypeDesc, buf interface{}) error

	Supports(tag string) bool
}

// Writer is implemented by image writers.
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

// NewWriter will attempt to create a writer for the given filename.  Type of image is
// deduced from the filename extension.
func NewWriter(filename string) (Writer, error) {
	for _, w := range writers {
		writer, err := w(filename)

		if writer != nil && err == nil {
			return writer, err
		}
	}

	return nil, errors.New("No Writer")
}

// Open will attempt to open an image and returns a Reader capable of loading it.  Callers should
// check the Reader for the capabilities of the image format.
func Open(filename string) (Reader, error) {
	for _, reader := range readers {

		r, err := reader(filename)

		if err == nil && r != nil {
			return r, err
		}

	}

	return nil, errors.New("image.Open: no reader found for " + filename)
}

// RegisterReader is called by the image reader libraries.
func RegisterReader(open func(filename string) (Reader, error)) {
	readers = append(readers, open)
}

// RegisterWriter is called by the image writer libraries.
func RegisterWriter(test func(filename string) (Writer, error)) {
	writers = append(writers, test)
}
