package tonemap

import (
	m "github.com/jamiec7919/vermeer/math"
)

// Tonemapping based on paper by Jiang Duan and Guoping Qiu
// http://www.cs.nott.ac.uk/~pszqiu/webpages/Papers/1494_Qiu_G.pdf

func Tonemap(w, h int, alpha float32, img []float32, buf []byte) error {

	H := make([]int32, 1000000)
	LI := make([]float32, w*h)

	Lmin := m.Inf(1)
	Lmax := m.Inf(-1)

	// Compute log luminance
	for i := range LI {
		LI[i] = m.Log(0.299*img[i*3+0] + 0.587*img[i*3+1] + 0.114*img[i*3+2])

		Lmin = m.Min(Lmin, LI[i])
		Lmax = m.Max(Lmax, LI[i])
	}

	// Compute histogram
	for i := range LI {
		l := (LI[i] - Lmin) / (Lmax - Lmin)

		i := int(l * float32(len(H)))

		if i > len(H)-1 {
			i = len(H) - 1
		}

		H[i]++
	}

	// Scan for beta0
	beta0 := 0
	sum := 0
	for sum < w*h/2 && beta0 < len(H) {
		sum += H[beta0]
		beta0++
	}

	dynrange := make([]float32, 256)
}

func rangedivide(hist []int32, Npix int) {
	beta := 0
	sum := 0
	for sum < Npix/2 && beta < len(hist) {
		sum += hist[beta]
		beta++
	}

	rangedivide(hist[:beta], Npix/2)
	rangedivide(hist[beta:], Npix/2)
}
