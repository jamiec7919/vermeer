// Copyright 2016 The Vermeer Light Tools Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hdr

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/jamiec7919/vermeer/image"
	"os"
)

type Reader struct {
	file   *os.File
	reader *bufio.Reader
	spec   image.Spec
}

func init() {

	open := func(filename string) (image.Reader, error) {
		r, err := Open(filename)

		if err == nil && r != nil {
			return r, err
		}
		return nil, err
	}

	image.RegisterReader(open)
}

func Open(filename string) (*Reader, error) {

	h := Reader{}

	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	h.file = file
	h.reader = bufio.NewReader(h.file)

	if err := h.readHeaders(); err != nil {
		return nil, err
	}

	return &h, nil
}

func (h *Reader) readHeaders() error {
	// bytes := make([]byte, DefaultBufferSize)

	for {

		line, err := h.reader.ReadString('\n')

		if err != nil {
			return err
		}
		if len(line) == 1 { // includes delim
			break
		}

		// log.Printf("%v", line)
	}

	line, err := h.reader.ReadString('\n')

	if err != nil {
		return err
	}

	var xs, ys string
	var width, height int
	fmt.Sscanf(line, "%s %d %s %d", &ys, &height, &xs, &width)

	// _,err := fmt.Fscanf...    include newline

	h.spec.Height = height
	h.spec.Width = width
	h.spec.X = 0 //xs
	h.spec.Y = 0 //ys
	h.spec.FullHeight = height
	h.spec.FullWidth = width
	h.spec.FullX = 0 //xs
	h.spec.FullY = 0 //ys
	h.spec.NChannels = 4
	h.spec.Format = []image.TypeDesc{image.TypeDesc{}, image.TypeDesc{}, image.TypeDesc{}, image.TypeDesc{}}
	h.spec.ChannelNames = []string{"R", "G", "B", "E"}
	h.spec.AlphaChannel = -1
	h.spec.ZChannel = -1

	return nil
}

func (h *Reader) Spec() (image.Spec, error) { return h.spec, nil }
func (h *Reader) Close()                    { h.file.Close() }

func (h *Reader) ReadImage(ty image.TypeDesc, buf interface{}) error {
	if ty.BaseType != image.FLOAT {
		return errors.New("HDR: only supports float32 pixels")
	}

	pbuf := buf.([]float32)

	if pbuf == nil {
		return errors.New("HDR: pixel buffer not float32")
	}

	scanline := make([]byte, h.spec.Width*4)

	for j := 0; j < h.spec.Height; j++ {
		if err := readScanline(h.reader, scanline[j*h.spec.Width*4:j*h.spec.Width*4+h.spec.Width*4]); err != nil {
			return err
		}

		for i := 0; i < h.spec.Width; i++ {
			or := scanline[(i*4)+0]
			og := scanline[(i*4)+1]
			ob := scanline[(i*4)+2]
			oe := scanline[(i*4)+3]

			expo := uint(oe) - 128

			r := convertComponent(expo, or)
			g := convertComponent(expo, og)
			b := convertComponent(expo, ob)

			pbuf[j*h.spec.Width*3+(i*3)+0] = r
			pbuf[j*h.spec.Width*3+(i*3)+1] = g
			pbuf[j*h.spec.Width*3+(i*3)+2] = b
		}
	}

	return nil
}

func (h *Reader) ReadScanline(y, z int, ty image.TypeDesc, buf interface{}) error {
	if ty.BaseType != image.FLOAT {
		return errors.New("HDR: only supports float32 pixels")
	}

	pbuf := buf.([]float32)

	if pbuf == nil {
		return errors.New("HDR: pixel buffer not float32")
	}

	//return nil
	return errors.New("HDR: scanline reading not supported")
}

func (h *Reader) ReadTile(x, y, z int, ty image.TypeDesc, buf interface{}) error {
	return errors.New("HDR: doesn't support tiles")
}

func (h *Reader) Supports(tag string) bool { return false }

func readScanlineOld(fin *bufio.Reader, scanline []byte) error {
	length := len(scanline) / 4

	s := make([]byte, 4)

	rshift := uint(0)
	l := 0

	for l < length {
		_, err := fin.Read(s)

		if err != nil {
			return err
		}

		if s[0] == 1 && s[1] == 1 && s[2] == 1 {
			// Encoded
			count := int(s[3]) << rshift
			//log.Printf("enc %v %v", l, length)
			for i := 0; i < count; i++ {
				scanline[(l+i)*4+0] = s[0]
				scanline[(l+i)*4+1] = s[1]
				scanline[(l+i)*4+2] = s[2]
				scanline[(l+i)*4+3] = s[3]
			}

			l += count
			rshift += 8
		} else {
			//log.Printf("nenc %v %v", l, length)
			scanline[l*4+0] = s[0]
			scanline[l*4+1] = s[1]
			scanline[l*4+2] = s[2]
			scanline[l*4+3] = s[3]
			l++
			rshift = 0
		}
	}
	return nil
}

func readScanline(fin *bufio.Reader, scanline []byte) error {
	l := len(scanline) / 4

	if l < minEncodingLen || l > maxEncodingLen {
		// log.Printf("MinMax")
		return readScanlineOld(fin, scanline)
	}

	shdr, err := fin.Peek(4)

	if err != nil {
		// log.Printf("shdr")
		return err
	}

	if shdr[0] != 2 || shdr[1] != 2 || (shdr[2]&128) != 0 {
		// Not new style after all
		// log.Printf("old")

		return readScanlineOld(fin, scanline)
	}

	if (int(shdr[2])<<8)|int(shdr[3]) != l {
		return nil //  Scanline length mismatch (should be error)
	}

	// Ok shdr is fine, advance
	fin.Read(shdr) // Can't fail since peek succeeded

	for i := 0; i < 4; i++ {
		j := 0

		for j < l {
			code, err := fin.ReadByte()

			if err != nil {
				return err
			}

			if code > 128 {
				code &= 127
				v, err := fin.ReadByte()

				if err != nil {
					return err
				}

				for k := 0; k < int(code); k++ {
					scanline[j*4+i] = v
					j++
				}
			} else {

				for k := 0; k < int(code); k++ {
					v, err := fin.ReadByte()

					if err != nil {
						return err
					}
					scanline[j*4+i] = v
					j++
				}
			}
		}
	}
	return nil
}
