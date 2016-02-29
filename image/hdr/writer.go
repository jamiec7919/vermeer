// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
	Package hdr implements implements the Radiance HDR format, supports only float32 RGB pixels.
*/
package hdr

import (
	"fmt"
	"os"
	//	"vermeer/image"
)

func writeScanline(fout *os.File, scanline []byte) error {
	fout.Write(scanline)
	return nil
}

func WriteHDR(filename string, w, h int, rgb []float32) error {
	fout, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fout.Close()

	fmt.Fprintf(fout, "#?RADIANCE\n")
	fmt.Fprintf(fout, "# %v\n", "Created by Vermeer Light Tools (http://www.vermeerlt.org)")
	fmt.Fprintf(fout, "FORMAT=32-bit_rle_rgbe\n")
	fmt.Fprintf(fout, "\n")
	fmt.Fprintf(fout, "+Y %v +X %v\n", h, w)

	scanline := make([]byte, w*4)

	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			k := h - j - 1
			r, g, b, e := RGBEOfRGB(rgb[(i+(k*w))*3+0], rgb[(i+(k*w))*3+1], rgb[(i+(k*w))*3+2])
			scanline[i*4+0] = r
			scanline[i*4+1] = g
			scanline[i*4+2] = b
			scanline[i*4+3] = e
		}
		writeScanline(fout, scanline)
	}
	return nil
}
