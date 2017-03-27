package texture

import (
	//"encoding/binary"
	"fmt"
	m "github.com/jamiec7919/vermeer/math"
	//"os"
)

type miplevel struct {
	w, h   int
	mipmap []byte
}

func (mip *miplevel) BilinearSample(s, t float32, cp int) (c [3]float32) {

	ms := s - m.Floor(s)
	mt := t - m.Floor(t)

	// ms and mt in [0,1)  (wrapped mode)

	x0 := int(m.Floor(ms * float32(mip.w)))
	x1 := int(m.Ceil(ms * float32(mip.w)))
	dx := ms*float32(mip.w) - m.Floor(ms*float32(mip.w))
	y0 := int(m.Floor(mt * float32(mip.h)))
	y1 := int(m.Ceil(mt * float32(mip.h)))
	dy := mt*float32(mip.h) - m.Floor(mt*float32(mip.h))

	x0 %= mip.w
	x1 %= mip.w

	if x0 < 0 {
		x0 += mip.w
	}

	if x1 < 0 {
		x1 += mip.w
	}

	y0 %= mip.h
	y1 %= mip.h

	if y0 < 0 {
		y0 += mip.h
	}
	if y1 < 0 {
		y1 += mip.h
	}

	for k := range c {
		c0 := (1-dx)*float32(mip.mipmap[(x0+(y0*mip.w))*cp+k]) + dx*float32(mip.mipmap[(x1+(y0*mip.w))*cp+k])
		c1 := (1-dx)*float32(mip.mipmap[(x0+(y1*mip.w))*cp+k]) + dx*float32(mip.mipmap[(x1+(y1*mip.w))*cp+k])

		c[k] = (1-dy)*c0 + dy*c1
	}

	return
}

type mipmap struct {
	components int
	mipmap     []miplevel
}

func (mip *mipmap) MaxLevelOfDetail() int {
	return len(mip.mipmap) - 1
}

func (mip *mipmap) TrilinearSample(s, t, lod float32) (c [3]float32) {
	l0 := int(m.Ceil(lod))
	l1 := int(m.Floor(lod))
	dl := lod - m.Floor(lod)

	if l0 < 0 {
		l0 = 0
	}

	if l0 > len(mip.mipmap)-1 {
		l0 = len(mip.mipmap) - 1
	}

	if l1 < 0 {
		l1 = 0
	}

	if l1 > len(mip.mipmap)-1 {
		l1 = len(mip.mipmap) - 1
	}

	if l1 == l0 {
		return mip.mipmap[l0].BilinearSample(s, t, mip.components)
	}

	c0 := mip.mipmap[l0].BilinearSample(s, t, mip.components)
	c1 := mip.mipmap[l1].BilinearSample(s, t, mip.components)

	for k := range c {
		c[k] = dl*c0[k] + (1-dl)*c1[k]
	}

	return
}

func maxi(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func mini(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var tex = 0

// http://http.download.nvidia.com/developer/Papers/2005/NP2_Mipmapping/NP2_Mipmap_Creation.pdf
// TODO: this might not be calculating down to 1x1 (maxlevel+1), double check indices
func stdfilter(w, h int, img []byte, components int) (out mipmap) {
	maxlevel := int(m.Ceil(m.Log2(m.Max(float32(w), float32(h)))))

	out.mipmap = make([]miplevel, maxlevel)

	out.mipmap[0].mipmap = img
	out.mipmap[0].w = w
	out.mipmap[0].h = h

	// wid[l] = max(1,floor(width>>1))
	width := w
	height := h

	for l := 1; l < maxlevel; l++ {
		nwidth := int(m.Max(1, m.Ceil(float32(width)/2)))
		nheight := int(m.Max(1, m.Ceil(float32(height)/2)))

		out.mipmap[l].mipmap = make([]byte, nwidth*nheight*components)
		out.mipmap[l].w = nwidth
		out.mipmap[l].h = nheight

		if nheight%2 == 0 {
			if nwidth%2 == 0 {
				for y := 0; y < nheight; y++ {
					y0 := y * 2
					y1 := mini(y0+1, maxi(1, height-1))

					for x := 0; x < nwidth; x++ {
						x0 := x * 2
						x1 := mini(x0+1, maxi(1, width-1))

						for k := 0; k < components; k++ {
							out.mipmap[l].mipmap[(x+y*nwidth)*components+k] = byte(0.25 * (float32(out.mipmap[l-1].mipmap[(x0+y0*width)*components+k]) +
								float32(out.mipmap[l-1].mipmap[(x0+y1*width)*components+k]) +
								float32(out.mipmap[l-1].mipmap[(x1+y0*width)*components+k]) +
								float32(out.mipmap[l-1].mipmap[(x1+y1*width)*components+k])))
						}
					}
				}
			} else {
				// height even, width odd
				for y := 0; y < nheight; y++ {
					y0 := y * 2
					y1 := mini(y0+1, maxi(1, height-1))

					for x := 0; x < nwidth; x++ {
						x0 := maxi(x*2-1, -maxi(1, width-1))
						x1 := x * 2
						x2 := mini(x*2+1, maxi(1, width-1))
						// wrap mode
						if x0 < 0 {
							x0 += width
							x0 = mini(x0, maxi(1, width-1))
						}
						if x2 > width-1 {
							x2 -= width
							x2 = maxi(x2, -maxi(1, width-1))
						}

						w0 := float32(nwidth-x-1) / float32(2*nwidth-1)
						w1 := float32(nwidth) / float32(2*nwidth-1)
						w2 := float32(x) / float32(2*nwidth-1)

						for k := 0; k < components; k++ {
							c00 := float32(out.mipmap[l-1].mipmap[(x0+y0*width)*components+k])
							c10 := float32(out.mipmap[l-1].mipmap[(x1+y0*width)*components+k])
							c20 := float32(out.mipmap[l-1].mipmap[(x2+y0*width)*components+k])
							c01 := float32(out.mipmap[l-1].mipmap[(x0+y1*width)*components+k])
							c11 := float32(out.mipmap[l-1].mipmap[(x1+y1*width)*components+k])
							c21 := float32(out.mipmap[l-1].mipmap[(x2+y1*width)*components+k])

							out.mipmap[l].mipmap[(x+y*nwidth)*components+k] = byte(0.5 * (w0*c00 + w1*c10 + w2*c20 + w0*c01 + w1*c11 + w2*c21))
						}
					}
				}
			}
		} else {
			if nwidth%2 == 0 {
				// height odd, width even
				for y := 0; y < nheight; y++ {
					y0 := maxi(y*2-1, -maxi(1, height-1))
					y1 := y * 2
					y2 := mini(y*2+1, maxi(1, height-1))
					// wrap mode
					if y0 < 0 {
						y0 += height
						y0 = mini(y0, maxi(1, height-1))

					}

					w0 := float32(nheight-y-1) / float32(2*nheight-1)
					w1 := float32(nheight) / float32(2*nheight-1)
					w2 := float32(y) / float32(2*nheight-1)

					for x := 0; x < nwidth; x++ {
						x0 := x * 2
						x1 := mini(x0+1, maxi(1, width-1))

						for k := 0; k < components; k++ {

							c00 := float32(out.mipmap[l-1].mipmap[(x0+y0*width)*components+k])
							c01 := float32(out.mipmap[l-1].mipmap[(x0+y1*width)*components+k])
							c02 := float32(out.mipmap[l-1].mipmap[(x0+y2*width)*components+k])
							c10 := float32(out.mipmap[l-1].mipmap[(x1+y0*width)*components+k])
							c11 := float32(out.mipmap[l-1].mipmap[(x1+y1*width)*components+k])
							c12 := float32(out.mipmap[l-1].mipmap[(x1+y2*width)*components+k])

							out.mipmap[l].mipmap[(x+y*nwidth)*components+k] = byte(0.5 * (w0*c00 + w1*c01 + w2*c02 + w0*c10 + w1*c11 + w2*c12))
						}
					}
				}
			} else {
				for y := 0; y < nheight; y++ {
					y0 := maxi(y*2-1, -maxi(1, height-1))
					y1 := y * 2
					y2 := mini(y*2+1, maxi(1, height-1))
					// wrap mode
					if y0 < 0 {
						y0 += height
						y0 = mini(y0, maxi(1, height-1))
					}

					wy0 := float32(nheight-y-1) / float32(2*nheight-1)
					wy1 := float32(nheight) / float32(2*nheight-1)
					wy2 := float32(y) / float32(2*nheight-1)

					for x := 0; x < nwidth; x++ {
						x0 := maxi(x*2-1, -maxi(1, width-1))
						x1 := x * 2
						x2 := mini(x*2+1, maxi(1, width-1))
						// wrap mode
						if x0 < 0 {
							x0 += width
							x0 = mini(x0, maxi(1, width-1))
						}
						if x2 > width-1 {
							x2 -= width
							x2 = maxi(x2, -maxi(1, width-1))
						}

						w0 := float32(nwidth-x-1) / float32(2*nwidth-1)
						w1 := float32(nwidth) / float32(2*nwidth-1)
						w2 := float32(x) / float32(2*nwidth-1)

						for k := 0; k < components; k++ {
							if false {
								fmt.Printf("%v %v %v %v %v\n", l, x0, y0, width, height)
							}
							c00 := float32(out.mipmap[l-1].mipmap[(x0+y0*width)*components+k])
							c01 := float32(out.mipmap[l-1].mipmap[(x0+y1*width)*components+k])
							c02 := float32(out.mipmap[l-1].mipmap[(x0+y2*width)*components+k])
							c10 := float32(out.mipmap[l-1].mipmap[(x1+y0*width)*components+k])
							c11 := float32(out.mipmap[l-1].mipmap[(x1+y1*width)*components+k])
							c12 := float32(out.mipmap[l-1].mipmap[(x1+y2*width)*components+k])
							c20 := float32(out.mipmap[l-1].mipmap[(x2+y0*width)*components+k])
							c21 := float32(out.mipmap[l-1].mipmap[(x2+y1*width)*components+k])
							c22 := float32(out.mipmap[l-1].mipmap[(x2+y2*width)*components+k])

							out.mipmap[l].mipmap[(x+y*nwidth)*components+k] = byte(wy0*(w0*c00+w1*c10+w2*c20) +
								wy1*(w0*c01+w1*c11+w2*c21) +
								wy2*(w0*c02+w1*c12+w2*c22))
						}
					}
				}

			}
		}
		/*
				buf := make([]float32, nwidth*nheight*components)

				for k := 0; k < nwidth*nheight*components; k++ {
					buf[k] = float32(out.mipmap[l].mipmap[k])
				}
				fp, err := os.Create(fmt.Sprintf("out%v_%v.float", tex, l))
				fmt.Printf("%v %v %v %v\n", fmt.Sprintf("out%v_%v.float", tex, l), nwidth, nheight, components)
				if err != nil {
					goto skip
				}
				err = binary.Write(fp, binary.LittleEndian, buf)
				if err != nil {
					goto skip
				}
			skip:
				fp.Close()
		*/
		width = nwidth
		height = nheight

	}
	tex++
	out.components = components

	return
}
