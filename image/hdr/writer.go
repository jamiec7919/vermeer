// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hdr

import (
	"errors"
	"fmt"
	"github.com/jamiec7919/vermeer/image"
	"os"
	"path/filepath"
)

type Writer struct {
	file *os.File
	spec image.Spec
}

func init() {
	image.RegisterWriter(func(filename string) (image.Writer, error) {
		ext := filepath.Ext(filename)

		if ext == ".hdr" {
			return &Writer{}, nil
		}

		return nil, image.ErrNoWriter
	})
}

func (w *Writer) Open(filename string, spec *image.Spec) error {
	return w.OpenMode(filename, spec, "")
}
func (w *Writer) OpenMode(filename string, spec *image.Spec, mode string) error {
	file, err := os.Create(filename)

	if err != nil {
		return err
	}

	w.file = file
	w.spec = *spec

	return nil
}

func (w *Writer) Close() {
	w.file.Close()
}

func (w *Writer) WriteImage(ty image.TypeDesc, buf interface{}) error {
	if ty.BaseType != image.FLOAT {
		return errors.New("HDR: only supports float32 pixels")
	}

	pbuf := buf.([]float32)

	if pbuf == nil {
		return errors.New("HDR: pixel buffer not float32")
	}

	fmt.Fprintf(w.file, "#?RADIANCE\n")
	fmt.Fprintf(w.file, "# %v\n", "Created by Vermeer Light Tools (http://www.vermeerlt.com)")
	fmt.Fprintf(w.file, "FORMAT=32-bit_rle_rgbe\n")
	fmt.Fprintf(w.file, "\n")
	fmt.Fprintf(w.file, "+Y %v +X %v\n", w.spec.Height, w.spec.Width)

	scanline := make([]byte, w.spec.Width*4)

	for j := 0; j < w.spec.Height; j++ {
		for i := 0; i < w.spec.Width; i++ {
			k := w.spec.Height - j - 1
			r, g, b, e := convertRGBToRGBE(pbuf[(i+(k*w.spec.Width))*3+0], pbuf[(i+(k*w.spec.Width))*3+1], pbuf[(i+(k*w.spec.Width))*3+2])
			scanline[i*4+0] = r
			scanline[i*4+1] = g
			scanline[i*4+2] = b
			scanline[i*4+3] = e
		}
		writeScanline(w.file, scanline)
	}

	return nil
}

func (w *Writer) WriteImageStride(ty image.TypeDesc, buf interface{}, xstride, ystride, zstride int) error {
	return errors.New("HDR WriteImageStride: unsupported")
}
func (w *Writer) WriteScanline(y, z int, ty image.TypeDesc, buf interface{}) error {
	return errors.New("HDR WriteScanline: unsupported")

}
func (w *Writer) WriteScanlineStride(y, z int, ty image.TypeDesc, buf interface{}, xstride, ystride, zstride int) error {
	return errors.New("HDR WriteScanlineStride: unsupported")

}
func (w *Writer) WriteTile(x, y, z int, ty image.TypeDesc, buf interface{}) error {
	return errors.New("HDR WriteTile: unsupported")

}
func (w *Writer) WriteTileStride(x, y, z int, ty image.TypeDesc, buf interface{}, xstride, ystride, zstride int) error {
	return errors.New("HDR WriteTileStride: unsupported")

}

func (w *Writer) Supports(tag string) bool {
	return false
}

func writeScanline(fout *os.File, scanline []byte) error {
	fout.Write(scanline)
	return nil
}
