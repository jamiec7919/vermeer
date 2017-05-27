package texture

import (
	//"encoding/binary"
	"fmt"
	m "github.com/jamiec7919/vermeer/math"
	//"os"
)

// WrapMode is the type given to the texture wrap modes.
type WrapMode int

const (
	WrapModeMirror WrapMode = iota
	WrapModeClamp
	WrapModeRepeat
	WrapModeBorder
)

type miplevel struct {
	w, h   int
	mipmap []byte
}

// TODO: Rewrite comment
// t is a value that goes from 0 to 1 to interpolate in a C1 continuous way across uniformly sampled data points.
// when t is 0, this will return B.  When t is 1, this will return C.  Inbetween values will return an interpolation
// between B and C.  A and B are used to calculate slopes at the edges.
func cubicHermite(A, B, C, D, t float32) float32 {
	a := -A/2.0 + (3.0*B)/2.0 - (3.0*C)/2.0 + D/2.0
	b := A - (5.0*B)/2.0 + 2.0*C - D/2.0
	c := -A/2.0 + C/2.0
	d := B

	return a*t*t*t + b*t*t + c*t + d
}

// BilcubicSample returns a (Hermite) bilcubic sampled colour from the 16 texels surrounding (s,t).
// The texture is in [0,1) but s and t can be outside this range.
func (mip *miplevel) BicubicSample(s, t float32, wrapS, wrapT WrapMode, cp int) (c [3]float32) {
	// take -1,0,1,2 in each direction
	xi := int(s)
	yi := int(t)

	//mirrorClampToEdge(f) = min(1-1/(2*N), max(1/(2*N), abs(f)))
	//mirrorClampToBorder(f) = min(1+1/(2*N), max(1/(2*N), abs(f)))
	ds := s - m.Floor(s)
	dt := t - m.Floor(t)

	switch wrapS {
	case WrapModeMirror:
		if xi&1 == 1 {
			s = 1 - ds
		} else {
			s = ds
		}
	case WrapModeClamp:
		if s < 0 {
			s = 0
		}
		if s > 1 {
			s = 1
		}
	default:
		s = ds
	}

	switch wrapT {
	case WrapModeMirror:
		if yi&1 == 1 {
			t = 1 - dt
		} else {
			t = dt
		}
	case WrapModeClamp:

		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
	default:
		t = dt
	}

	xf := s*float32(mip.w) - 0.5
	x0i := int(xf)
	xm1i := x0i - 1
	x1i := x0i + 1
	x2i := x0i + 2
	wx := xf - float32(x0i)

	if x0i < 0 {
		switch wrapS {
		case WrapModeMirror:
			x0i = 0
		case WrapModeRepeat:
			x0i = mip.w - 1
		default:
			x0i = 0
		}
	}

	if xm1i < 0 {
		switch wrapS {
		case WrapModeMirror:
			xm1i = -xm1i - 1
		case WrapModeRepeat:
			xm1i = mip.w - 1 + xm1i
		default:
			xm1i = 0
		}
	}

	if x1i > mip.w-1 {
		switch wrapS {
		case WrapModeMirror:
			x1i = mip.w - 1
		case WrapModeRepeat:
			x1i = 0
		default:
			x1i = mip.w - 1
		}
	}

	if x2i > mip.w-1 {
		switch wrapS {
		case WrapModeMirror:
			x2i = 2*mip.w - x2i - 1
		case WrapModeRepeat:
			x2i = x2i - mip.w
		default:
			x2i = mip.w - 1
		}
	}

	yf := t*float32(mip.h) - 0.5
	y0i := int(yf)
	ym1i := y0i - 1
	y1i := y0i + 1
	y2i := y0i + 2
	wy := yf - float32(y0i)

	if y0i < 0 {
		switch wrapT {
		case WrapModeMirror:
			y0i = 0
		case WrapModeRepeat:
			y0i = mip.h - 1
		default:
			y0i = 0
		}
	}

	if ym1i < 0 {
		switch wrapT {
		case WrapModeMirror:
			ym1i = -ym1i - 1
		case WrapModeRepeat:
			ym1i = mip.h - 1 + ym1i
		default:
			ym1i = 0
		}
	}

	if y1i > mip.h-1 {
		switch wrapT {
		case WrapModeMirror:
			y1i = mip.h - 1
		case WrapModeRepeat:
			y1i = 0
		default:
			y1i = mip.h - 1
		}
	}

	if y2i > mip.h-1 {
		switch wrapT {
		case WrapModeMirror:
			y2i = 2*mip.h - y2i - 1
		case WrapModeRepeat:
			y2i = y2i - mip.h
		default:
			y2i = mip.h - 1
		}
	}

	if x0i < 0 || xm1i < 0 ||
		x1i < 0 || x2i < 0 ||
		x0i > mip.w-1 || xm1i > mip.w-1 ||
		x1i > mip.w-1 || x2i > mip.w-1 ||
		y0i < 0 || ym1i < 0 ||
		y1i < 0 || y2i < 0 ||
		y0i > mip.h-1 || ym1i > mip.h-1 ||
		y1i > mip.h-1 || y2i > mip.h-1 {
		fmt.Printf("%v %v %v %v, %v %v %v %v (%v,%v)", xm1i, x0i, x1i, x2i, ym1i, y0i, y1i, y2i, mip.w, mip.h)
	}

	for k := range c {
		p00 := float32(mip.mipmap[(xm1i+(ym1i*mip.w))*cp+k])
		p10 := float32(mip.mipmap[(x0i+(ym1i*mip.w))*cp+k])
		p20 := float32(mip.mipmap[(x1i+(ym1i*mip.w))*cp+k])
		p30 := float32(mip.mipmap[(x2i+(ym1i*mip.w))*cp+k])

		p01 := float32(mip.mipmap[(xm1i+(y0i*mip.w))*cp+k])
		p11 := float32(mip.mipmap[(x0i+(y0i*mip.w))*cp+k])
		p21 := float32(mip.mipmap[(x1i+(y0i*mip.w))*cp+k])
		p31 := float32(mip.mipmap[(x2i+(y0i*mip.w))*cp+k])

		p02 := float32(mip.mipmap[(xm1i+(y1i*mip.w))*cp+k])
		p12 := float32(mip.mipmap[(x0i+(y1i*mip.w))*cp+k])
		p22 := float32(mip.mipmap[(x1i+(y1i*mip.w))*cp+k])
		p32 := float32(mip.mipmap[(x2i+(y1i*mip.w))*cp+k])

		p03 := float32(mip.mipmap[(xm1i+(y2i*mip.w))*cp+k])
		p13 := float32(mip.mipmap[(x0i+(y2i*mip.w))*cp+k])
		p23 := float32(mip.mipmap[(x1i+(y2i*mip.w))*cp+k])
		p33 := float32(mip.mipmap[(x2i+(y2i*mip.w))*cp+k])

		c0 := cubicHermite(p00, p10, p20, p30, wx)
		c1 := cubicHermite(p01, p11, p21, p31, wx)
		c2 := cubicHermite(p02, p12, p22, p32, wx)
		c3 := cubicHermite(p03, p13, p23, p33, wx)

		cc := cubicHermite(c0, c1, c2, c3, wy)

		if cc < 0 {
			cc = 0
		}
		if cc > 255 {
			cc = 255
		}

		c[k] = cc
	}

	return c
}

// BilinearSample2 returns a bilinearly sampled colour from the 4 texels surrounding (s,t).
// The texture is in [0,1) but s and t can be outside this range.
func (mip *miplevel) BilinearSample2(s, t float32, wrapS, wrapT WrapMode, cp int) (c [3]float32) {

	xi := int(s)
	yi := int(t)

	//mirrorClampToEdge(f) = min(1-1/(2*N), max(1/(2*N), abs(f)))
	//mirrorClampToBorder(f) = min(1+1/(2*N), max(1/(2*N), abs(f)))
	ds := s - m.Floor(s)
	dt := t - m.Floor(t)

	switch wrapS {
	case WrapModeMirror:
		if xi&1 == 1 {
			s = 1 - ds
		} else {
			s = ds
		}
	case WrapModeClamp:
		if s < 0 {
			s = 0
		}
		if s > 1 {
			s = 1
		}
	default:
		s = ds
	}

	switch wrapT {
	case WrapModeMirror:
		if yi&1 == 1 {
			t = 1 - dt
		} else {
			t = dt
		}
	case WrapModeClamp:

		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
	default:
		t = dt
	}

	xf := s*float32(mip.w) - 0.5
	x0i := int(xf)
	x1i := x0i + 1
	wx := xf - float32(x0i)

	if x0i < 0 {
		switch wrapS {
		case WrapModeMirror:
			x0i = 0
		case WrapModeRepeat:
			x0i = mip.w - 1
		default:
			x0i = 0
		}
	}

	if x1i > mip.w-1 {
		switch wrapS {
		case WrapModeMirror:
			x1i = mip.w - 1
		case WrapModeRepeat:
			x1i = 0
		default:
			x1i = mip.w - 1
		}
	}

	yf := t*float32(mip.h) - 0.5
	y0i := int(yf)
	y1i := y0i + 1
	wy := yf - float32(y0i)

	if y0i < 0 {
		switch wrapT {
		case WrapModeMirror:
			y0i = 0
		case WrapModeRepeat:
			y0i = mip.h - 1
		default:
			y0i = 0
		}
	}

	if y1i > mip.h-1 {
		switch wrapT {
		case WrapModeMirror:
			y1i = mip.h - 1
		case WrapModeRepeat:
			y1i = 0
		default:
			y1i = mip.h - 1
		}
	}

	for k := range c {
		c0 := (1-wx)*float32(mip.mipmap[(x0i+(y0i*mip.w))*cp+k]) + wx*float32(mip.mipmap[(x1i+(y0i*mip.w))*cp+k])
		c1 := (1-wx)*float32(mip.mipmap[(x0i+(y1i*mip.w))*cp+k]) + wx*float32(mip.mipmap[(x1i+(y1i*mip.w))*cp+k])

		cc := (1-wy)*c0 + wy*c1

		if cc < 0 {
			cc = 0
		}
		if cc > 255 {
			cc = 255
		}

		c[k] = cc
	}

	return c
}

// BilinearSample returns a bilinearly sampled colour from the 4 texels surrounding (s,t).
// The texture is in [0,1) but s and t can be outside this range.
func (mip *miplevel) BilinearSample(s, t float32, wrapS, wrapT WrapMode, cp int) (c [3]float32) {
	return mip.BilinearSample2(s, t, wrapS, wrapT, cp)
	ms := s - m.Floor(s)
	mt := t - m.Floor(t)

	if wrapS == WrapModeMirror {
		if s < 0 {
			s = -s
		}
		ms = s - m.Floor(s)
		//		ms = 1 - ms
	}

	if wrapT == WrapModeMirror {
		if t < 0 {
			t = -t
		}
		mt = t - m.Floor(t)
		//		mt = 1 - mt
	}

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

func (mip *mipmap) TrilinearSample(s, t, lod float32, wrapS, wrapT WrapMode) (c [3]float32) {
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
		if l0 == 0 {
			return mip.mipmap[l0].BicubicSample(s, t, wrapS, wrapT, mip.components)
			//return mip.mipmap[l0].BilinearSample(s, t, wrapS, wrapT, mip.components)
		} else {
			return mip.mipmap[l0].BilinearSample(s, t, wrapS, wrapT, mip.components)

		}
	}

	c0 := mip.mipmap[l0].BilinearSample(s, t, wrapS, wrapT, mip.components)
	c1 := mip.mipmap[l1].BilinearSample(s, t, wrapS, wrapT, mip.components)

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

// Round down
func AverageMipMapFilterRU(w, h, components int, img []uint8) mipmap {

	maxLevel := int(m.Ceil(m.Log2(m.Max(float32(w), float32(h)))))

	var out mipmap

	out.mipmap = make([]miplevel, maxLevel)

	out.mipmap[0].mipmap = img
	out.mipmap[0].w = w
	out.mipmap[0].h = h

	width := w
	height := h

	for l := 1; l < maxLevel; l++ {
		newWidth := int(m.Ceil(float32(width) / 2))   // Round up
		newHeight := int(m.Ceil(float32(height) / 2)) // Round up

		out.mipmap[l].mipmap = make([]byte, newWidth*newHeight*components)
		out.mipmap[l].w = newWidth
		out.mipmap[l].h = newHeight

		// Now depends on if old is odd or even
		horizIsEven := width&1 == 0
		vertIsEven := height&1 == 0

		switch {
		case horizIsEven && vertIsEven:
			for y := 0; y < newHeight; y++ {
				y0 := y * 2
				y1 := y*2 + 1

				for x := 0; x < newWidth; x++ {
					x0 := x * 2
					x1 := x*2 + 1

					for k := 0; k < components; k++ {
						out.mipmap[l].mipmap[(x+y*newWidth)*components+k] = uint8(0.25 * (float32(out.mipmap[l-1].mipmap[(x0+y0*width)*components+k]) +
							float32(out.mipmap[l-1].mipmap[(x0+y1*width)*components+k]) +
							float32(out.mipmap[l-1].mipmap[(x1+y0*width)*components+k]) +
							float32(out.mipmap[l-1].mipmap[(x1+y1*width)*components+k])))
					}

				}
			}
		case horizIsEven && !vertIsEven:
			for y := 0; y < newHeight; y++ {
				y0 := y*2 - 1
				y1 := y * 2
				y2 := y*2 + 1

				wnorm := float32(2*newHeight - 1)    // Round down
				w0 := float32(newHeight-y-1) / wnorm // if y == 0 then w0 == (newHeight-1)/(2*newHeight-1)
				w1 := float32(newHeight) / wnorm
				w2 := float32(y) / wnorm

				// These wrap modes depend..,  could clamp (to 0) or mirror.
				if y0 < 0 {
					//w0 = 0.0     // Clamp to 0.0
					//y0 += height // Wrap
					y0 = 0 // Mirror
				}
				if y2 > height-1 {
					//w2 = 0.0     // Clamp to 0.0
					//y2 -= height // Wrap
					y2 = height - 1 // Mirror
				}

				for x := 0; x < newWidth; x++ {
					x0 := x * 2
					x1 := x*2 + 1

					for k := 0; k < components; k++ {

						c00 := float32(out.mipmap[l-1].mipmap[(x0+y0*width)*components+k])
						c01 := float32(out.mipmap[l-1].mipmap[(x0+y1*width)*components+k])
						c02 := float32(out.mipmap[l-1].mipmap[(x0+y2*width)*components+k])
						c10 := float32(out.mipmap[l-1].mipmap[(x1+y0*width)*components+k])
						c11 := float32(out.mipmap[l-1].mipmap[(x1+y1*width)*components+k])
						c12 := float32(out.mipmap[l-1].mipmap[(x1+y2*width)*components+k])

						out.mipmap[l].mipmap[(x+y*newWidth)*components+k] = byte(0.5 * (w0*c00 + w1*c01 + w2*c02 + w0*c10 + w1*c11 + w2*c12))
					}

				}
			}
		case !horizIsEven && vertIsEven:
			for y := 0; y < newHeight; y++ {
				y0 := y * 2
				y1 := y*2 + 1

				for x := 0; x < newWidth; x++ {
					x0 := x*2 - 1
					x1 := x * 2
					x2 := x*2 + 1

					wnorm := float32(2*newWidth - 1)    // Round down
					w0 := float32(newWidth-x-1) / wnorm // if y == 0 then w0 == (newHeight-1)/(2*newHeight-1)
					w1 := float32(newWidth) / wnorm
					w2 := float32(x) / wnorm

					// These wrap modes depend..,  could clamp (to 0) or mirror.
					if x0 < 0 {
						//w0 = 0.0    // Clamp to 0.0
						//x0 += width // Wrap
						x0 = 0 // Mirror
					}
					if x2 > width-1 {
						//w2 = 0.0    // Clamp to 0.0
						//x2 -= width // Wrap
						x2 = width - 1 // Mirror
					}

					for k := 0; k < components; k++ {
						c00 := float32(out.mipmap[l-1].mipmap[(x0+y0*width)*components+k])
						c10 := float32(out.mipmap[l-1].mipmap[(x1+y0*width)*components+k])
						c20 := float32(out.mipmap[l-1].mipmap[(x2+y0*width)*components+k])
						c01 := float32(out.mipmap[l-1].mipmap[(x0+y1*width)*components+k])
						c11 := float32(out.mipmap[l-1].mipmap[(x1+y1*width)*components+k])
						c21 := float32(out.mipmap[l-1].mipmap[(x2+y1*width)*components+k])

						out.mipmap[l].mipmap[(x+y*newWidth)*components+k] = byte(0.5 * (w0*c00 + w1*c10 + w2*c20 + w0*c01 + w1*c11 + w2*c21))
					}

				}
			}
		default: // Both odd
			for y := 0; y < newHeight; y++ {
				y0 := y*2 - 1
				y1 := y * 2
				y2 := y*2 + 1

				wynorm := float32(2*newHeight - 1)     // Round down
				wy0 := float32(newHeight-y-1) / wynorm // if y == 0 then w0 == (newHeight-1)/(2*newHeight-1)
				wy1 := float32(newHeight) / wynorm
				wy2 := float32(y) / wynorm

				// These wrap modes depend..,  could clamp (to 0) or mirror.
				if y0 < 0 {
					//wy0 = 0.0    // Clamp to 0.0
					//y0 += height // Wrap
					y0 = 0 // Mirror
				}
				if y2 > height-1 {
					//wy2 = 0.0    // Clamp to 0.0
					//y2 -= height // Wrap
					y2 = height - 1 // Mirror
				}

				for x := 0; x < newWidth; x++ {
					x0 := x*2 - 1
					x1 := x * 2
					x2 := x*2 + 1

					wxnorm := float32(2*newWidth - 1)     // Round down
					wx0 := float32(newWidth-x-1) / wxnorm // if y == 0 then w0 == (newHeight-1)/(2*newHeight-1)
					wx1 := float32(newWidth) / wxnorm
					wx2 := float32(x) / wxnorm

					// These wrap modes depend..,  could clamp (to 0) or mirror.
					if x0 < 0 {
						//wx0 = 0.0   // Clamp to 0.0
						//x0 += width // Wrap
						x0 = 0 // Mirror
					}
					if x2 > width-1 {
						//wx2 = 0.0   // Clamp to 0.0
						//x2 -= width // Wrap
						x2 = width - 1 // Mirror
					}

					for k := 0; k < components; k++ {
						c00 := float32(out.mipmap[l-1].mipmap[(x0+y0*width)*components+k])
						c10 := float32(out.mipmap[l-1].mipmap[(x1+y0*width)*components+k])
						c20 := float32(out.mipmap[l-1].mipmap[(x2+y0*width)*components+k])
						c01 := float32(out.mipmap[l-1].mipmap[(x0+y1*width)*components+k])
						c11 := float32(out.mipmap[l-1].mipmap[(x1+y1*width)*components+k])
						c21 := float32(out.mipmap[l-1].mipmap[(x2+y1*width)*components+k])
						c02 := float32(out.mipmap[l-1].mipmap[(x0+y2*width)*components+k])
						c12 := float32(out.mipmap[l-1].mipmap[(x1+y2*width)*components+k])
						c22 := float32(out.mipmap[l-1].mipmap[(x2+y2*width)*components+k])

						out.mipmap[l].mipmap[(x+y*newWidth)*components+k] = byte(wx0*wy0*c00 + wx1*wy0*c10 + wx2*wy0*c20 +
							wx0*wy1*c01 + wx1*wy1*c11 + wx2*wy0*c21 +
							wx0*wy2*c02 + wx1*wy2*c12 + wx2*wy0*c22)
					}

				}
			}

		}
		width = newWidth
		height = newHeight

	}

	tex++
	out.components = components

	return out

}

// Round down
func AverageMipMapFilterRD(w, h, components int, img []uint8) mipmap {

	maxLevel := int(m.Ceil(m.Log2(m.Max(float32(w), float32(h)))))

	var out mipmap

	out.mipmap = make([]miplevel, maxLevel)

	out.mipmap[0].mipmap = img
	out.mipmap[0].w = w
	out.mipmap[0].h = h

	width := w
	height := h

	for l := 1; l < maxLevel; l++ {
		newWidth := width / 2   // Round down
		newHeight := height / 2 // Round down

		out.mipmap[l].mipmap = make([]byte, newWidth*newHeight*components)
		out.mipmap[l].w = newWidth
		out.mipmap[l].h = newHeight

		// Now depends on if old is odd or even
		horizIsEven := width&1 == 0
		vertIsEven := height&1 == 0

		switch {
		case horizIsEven && vertIsEven:
			for y := 0; y < newHeight; y++ {
				y0 := y * 2
				y1 := mini(y*2+1, height-1)

				for x := 0; x < newWidth; x++ {
					x0 := x * 2
					x1 := mini(x*2+1, width-1)

					for k := 0; k < components; k++ {
						out.mipmap[l].mipmap[(x+y*newWidth)*components+k] = uint8(0.25 * (float32(out.mipmap[l-1].mipmap[(x0+y0*width)*components+k]) +
							float32(out.mipmap[l-1].mipmap[(x0+y1*width)*components+k]) +
							float32(out.mipmap[l-1].mipmap[(x1+y0*width)*components+k]) +
							float32(out.mipmap[l-1].mipmap[(x1+y1*width)*components+k])))
					}

				}
			}
		case horizIsEven && !vertIsEven:
			for y := 0; y < newHeight; y++ {
				y0 := y * 2
				y1 := mini(y*2+1, height-1)
				y2 := mini(y*2+2, height-1)

				wnorm := float32(2*newHeight + 1) // Round down
				w0 := float32(newHeight-y) / wnorm
				w1 := float32(newHeight) / wnorm
				w2 := float32(y) / wnorm

				for x := 0; x < newWidth; x++ {
					x0 := x * 2
					x1 := mini(x*2+1, width-1)

					for k := 0; k < components; k++ {

						c00 := float32(out.mipmap[l-1].mipmap[(x0+y0*width)*components+k])
						c01 := float32(out.mipmap[l-1].mipmap[(x0+y1*width)*components+k])
						c02 := float32(out.mipmap[l-1].mipmap[(x0+y2*width)*components+k])
						c10 := float32(out.mipmap[l-1].mipmap[(x1+y0*width)*components+k])
						c11 := float32(out.mipmap[l-1].mipmap[(x1+y1*width)*components+k])
						c12 := float32(out.mipmap[l-1].mipmap[(x1+y2*width)*components+k])

						out.mipmap[l].mipmap[(x+y*newWidth)*components+k] = byte(0.5 * (w0*c00 + w1*c01 + w2*c02 + w0*c10 + w1*c11 + w2*c12))
					}

				}
			}
		case !horizIsEven && vertIsEven:
		default: // Both odd

		}

		width = newWidth
		height = newHeight

	}

	tex++
	out.components = components

	return out

}

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
					y1 := mini(y0+1, maxi(1, height-1)) // height always >= 1 ...

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
					y1 := mini(y*2, maxi(1, height-1))
					y2 := mini(y*2+1, height-1) // if height==1 this will fail
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

							if y2 > height-1 || y2 < 0 {
								fmt.Printf("%v %v", x0, y2)
							}
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
