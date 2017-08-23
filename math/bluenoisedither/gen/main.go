package main

/* This utility generates the blue-noise dithering patterns used as scrambles in Vermeer.
Currnetly it's very naive and outputs a fixed API.  Takes quite a while to generate good
patterns, giving it at least 1,000,000 iterations.
*/
import (
	//"encoding/binary"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"text/template"
)

var file = `package {{.Package}}

const TileSize = {{.TileSize}}

var Time1D = [{{.TileSize}}*{{.TileSize}}]uint64 {
	{{range .Time1D}}{{.}},
	{{end}}
}

var Lens2D = [{{.TileSize}}*{{.TileSize}}][2]uint64 {
	{{range .Lens2D}}{ {{index . 0}}, {{index . 1}} },
	{{end}}
}

`

var tpl = template.Must(template.New("file").Parse(file))

func Float64ToScramble(x float64) uint64 {
	bits := math.Float64bits(x + 1.0)

	return bits << (64 - 53)
}

// This calculates the total energy due to 1 pixel
func E1DPixel(size int, buf []float64, x, y int) float64 {
	Etotal := 0.0

	samp2 := buf[x+size*y]
	x2 := float64(x)
	y2 := float64(y)

	for j := 0; j < size; j++ {
		for i := 0; i < size; i++ {

			if i == x && j == y {
				continue
			}

			x1 := float64(i)
			y1 := float64(j)
			samp1 := buf[i+size*j]

			dx := x2 - x1
			dy := y2 - y1
			// Wrap boundary
			if true {
				if (x2+float64(size))-x1 < dx {
					dx = (x2 + float64(size)) - x1
				}
				if (y2+float64(size))-y1 < dy {
					dy = (y2 + float64(size)) - y1
				}
			}
			ispace := ((dx * dx) + (dy * dy)) / (2.1 * 2.1)

			dsamp := samp2 - samp1

			// ||ps-qs||^1/2
			isamp := math.Sqrt(math.Abs(dsamp)) / (1 * 1)

			E := math.Exp(-ispace - isamp)

			//log.Printf("%v: %v %v %v\n", E, ispace, isamp, dsamp)

			Etotal += E

		}
	}

	return Etotal
}

// This calculates the total energy due to 1 pixel
func E2DPixel(size int, buf []float64, x, y int) float64 {
	Etotal := 0.0

	samp21 := buf[(x+size*y)*2+0]
	samp22 := buf[(x+size*y)*2+1]
	x2 := float64(x)
	y2 := float64(y)

	for j := 0; j < size; j++ {
		for i := 0; i < size; i++ {

			if i == x && j == y {
				continue
			}

			x1 := float64(i)
			y1 := float64(j)
			samp11 := buf[(i+size*j)*2+0]
			samp12 := buf[(i+size*j)*2+1]

			dx := x2 - x1
			dy := y2 - y1
			// Wrap boundary
			if true {
				if (x2+float64(size))-x1 < dx {
					dx = (x2 + float64(size)) - x1
				}
				if (y2+float64(size))-y1 < dy {
					dy = (y2 + float64(size)) - y1
				}
			}
			ispace := ((dx * dx) + (dy * dy)) / (2.1 * 2.1)

			dsamp := (samp21-samp11)*(samp21-samp11) + (samp22-samp12)*(samp22-samp12)

			// ||ps-qs||^2/2
			isamp := math.Sqrt(dsamp) / (1 * 1)

			E := math.Exp(-ispace - isamp)

			//log.Printf("%v: %v %v %v\n", E, ispace, isamp, dsamp)

			Etotal += E

		}
	}

	return Etotal
}

func E1D(size int, buf []float64) float64 {
	Etotal := 0.0

	for j := 0; j < size; j++ {
		for i := 0; i < size; i++ {

			x1 := float64(i)
			y1 := float64(j)
			samp1 := buf[i+size*j]

			// Do rest of row first
			for i2 := i + 1; i2 < size; i2++ {

				x2 := float64(i2)
				y2 := float64(j)
				samp2 := buf[i2+size*j]

				dx := x2 - x1
				dy := y2 - y1

				// wrap boundary
				if true {
					if (x2+float64(size))-x1 < dx {
						dx = (x2 + float64(size)) - x1
					}
					if (y2+float64(size))-y1 < dy {
						dy = (y2 + float64(size)) - y1
					}
				}

				ispace := ((dx * dx) + (dy * dy)) / (2.1 * 2.1)

				dsamp := samp2 - samp1

				// ||ps-qs||^1/2
				isamp := math.Sqrt(math.Abs(dsamp)) / (1 * 1)

				E := math.Exp(-ispace - isamp)

				//log.Printf("%v: %v %v %v\n", E, ispace, isamp, dsamp)

				Etotal += E
			}

			// Now rest of image
			for j2 := j + 1; j2 < size; j2++ {
				for i2 := 0; i2 < size; i2++ {

					x2 := float64(i2)
					y2 := float64(j2)
					samp2 := buf[i2+size*j2]

					dx := x2 - x1
					dy := y2 - y1

					ispace := ((dx * dx) + (dy * dy)) / (2.1 * 2.1)

					dsamp := samp2 - samp1

					// ||ps-qs||^1/2
					isamp := math.Sqrt(math.Abs(dsamp)) / (1 * 1)

					Etotal += math.Exp(-ispace - isamp)

				}

			}

		}
	}

	return Etotal
}

func E2D(size int, buf []float64) float64 {
	Etotal := 0.0

	for j := 0; j < size; j++ {
		for i := 0; i < size; i++ {

			x1 := float64(i)
			y1 := float64(j)
			samp11 := buf[(i+size*j)*2+0]
			samp12 := buf[(i+size*j)*2+1]

			// Do rest of row first
			for i2 := i + 1; i2 < size; i2++ {

				x2 := float64(i2)
				y2 := float64(j)
				samp21 := buf[(i2+size*j)+0]
				samp22 := buf[(i2+size*j)+1]

				dx := x2 - x1
				dy := y2 - y1

				// wrap boundary
				if true {
					if (x2+float64(size))-x1 < dx {
						dx = (x2 + float64(size)) - x1
					}
					if (y2+float64(size))-y1 < dy {
						dy = (y2 + float64(size)) - y1
					}
				}

				ispace := ((dx * dx) + (dy * dy)) / (2.1 * 2.1)

				dsamp := (samp21-samp11)*(samp21-samp11) + (samp22-samp12)*(samp22-samp12)

				// ||ps-qs||^2/2
				isamp := math.Sqrt(dsamp) / (1 * 1)

				E := math.Exp(-ispace - isamp)

				//log.Printf("%v: %v %v %v\n", E, ispace, isamp, dsamp)

				Etotal += E
			}

			// Now rest of image
			for j2 := j + 1; j2 < size; j2++ {
				for i2 := 0; i2 < size; i2++ {

					x2 := float64(i2)
					y2 := float64(j2)
					samp21 := buf[(i2+size*j2)+0]
					samp22 := buf[(i2+size*j2)+1]

					dx := x2 - x1
					dy := y2 - y1

					ispace := ((dx * dx) + (dy * dy)) / (2.1 * 2.1)

					dsamp := (samp21-samp11)*(samp21-samp11) + (samp22-samp12)*(samp22-samp12)

					// ||ps-qs||^2/2
					isamp := math.Sqrt(dsamp) / (1 * 1)

					Etotal += math.Exp(-ispace - isamp)

				}

			}

		}
	}

	return Etotal
}

// size is size of tile to generate (e.g. 128x128) buf if size*size*1
func Blue1D(size, iter int, buf []float64) error {
	Estart := E1D(size, buf)

	for i := 0; i < iter; i++ {
		//Etotal := E1D(size, buf)

	retry:

		i1 := rand.Intn(size * size)
		i2 := rand.Intn(size * size)

		if i1 == i2 {
			goto retry
		}

		Etotal := E1DPixel(size, buf, i1%size, i1/size) + E1DPixel(size, buf, i2%size, i2/size)
		// test swap
		buf[i1], buf[i2] = buf[i2], buf[i1]

		Etest := E1DPixel(size, buf, i1%size, i1/size) + E1DPixel(size, buf, i2%size, i2/size)

		if Etest > Etotal {
			// nope, swap back
			buf[i1], buf[i2] = buf[i2], buf[i1]

		}

		/*
			ix1 := int(x1)
			iy1 := int(y1)
			ix2 := int(x2)
			iy2 := int(y2)

			dx := x2 - x1
			dy := y2 - y1

			ispace := ((dx * dx) + (dy * dy)) / (2.1 * 2.1)

			samp1 := buf[ix1+size*iy1]
			samp2 := buf[ix2+size*iy2]

			dsamp := samp2 - samp1

			// ||ps-qs||^1/2
			isamp := math.Sqrt(dsamp) / (1 * 1)

			E := math.Exp(-ispace - isamp)
		*/

		log.Printf("Iter %v %v %v\n", i, Etotal, Etest)
	}

	Eend := E1D(size, buf)

	log.Printf("End %v %v\n", Estart, Eend)

	return nil

}

// size is size of tile to generate (e.g. 128x128) buf if size*size*1
func Blue2D(size, iter int, buf []float64) error {
	Estart := E2D(size, buf)

	for i := 0; i < iter; i++ {
		//Etotal := E1D(size, buf)

	retry:

		i1 := rand.Intn(size * size)
		i2 := rand.Intn(size * size)

		if i1 == i2 {
			goto retry
		}

		Etotal := E2DPixel(size, buf, i1%size, i1/size) + E2DPixel(size, buf, i2%size, i2/size)
		// test swap
		buf[i1*2+0], buf[i2*2+0] = buf[i2*2+0], buf[i1*2+0]
		buf[i1*2+1], buf[i2*2+1] = buf[i2*2+1], buf[i1*2+1]

		Etest := E2DPixel(size, buf, i1%size, i1/size) + E2DPixel(size, buf, i2%size, i2/size)

		if Etest > Etotal {
			// nope, swap back
			buf[i1*2+0], buf[i2*2+0] = buf[i2*2+0], buf[i1*2+0]
			buf[i1*2+1], buf[i2*2+1] = buf[i2*2+1], buf[i1*2+1]

		}

		/*
			ix1 := int(x1)
			iy1 := int(y1)
			ix2 := int(x2)
			iy2 := int(y2)

			dx := x2 - x1
			dy := y2 - y1

			ispace := ((dx * dx) + (dy * dy)) / (2.1 * 2.1)

			samp1 := buf[ix1+size*iy1]
			samp2 := buf[ix2+size*iy2]

			dsamp := samp2 - samp1

			// ||ps-qs||^1/2
			isamp := math.Sqrt(dsamp) / (1 * 1)

			E := math.Exp(-ispace - isamp)
		*/

		log.Printf("Iter %v %v %v\n", i, Etotal, Etest)
	}

	Eend := E2D(size, buf)

	log.Printf("End %v %v\n", Estart, Eend)

	return nil

}

func White1D(size int, buf []float64) {
	for i := 0; i < size*size; i++ {
		buf[i] = rand.Float64()
		//buf[i] = float64(i%256) / 256.0
		//log.Printf("buf[%v] %v", i, buf[i])
	}
}

func White2D(size int, buf []float64) {
	for i := 0; i < size*size; i++ {
		buf[i*2] = rand.Float64()
		buf[i*2+1] = rand.Float64()
		//buf[i] = float64(i%256) / 256.0
		//log.Printf("buf[%v] %v", i, buf[i])
	}
}

func main() {
	rand.Seed(123134)
	size := 128
	/*
		buf := make([]float64, size*size)
		buf2 := make([]float32, size*size*3)

		White1D(size, buf)

		for i := range buf {
			buf2[i*3] = float32(buf[i] * 100.0)   //* 256
			buf2[i*3+1] = float32(buf[i] * 100.0) //* 256
			buf2[i*3+2] = float32(buf[i] * 100.0) //* 256
		}

		fp, err := os.Create("white.float")
		if err != nil {
			log.Printf("ERR: %v", err)
			return
		}
		err = binary.Write(fp, binary.LittleEndian, buf2)
		if err != nil {
			log.Printf("ERR: %v", err)
			return
		}
		fp.Close()

		Blue1D(size, 1000, buf)

		for i := range buf {
			buf2[i*3] = float32(buf[i] * 100.0)   //* 256
			buf2[i*3+1] = float32(buf[i] * 100.0) //* 256
			buf2[i*3+2] = float32(buf[i] * 100.0) //* 256
		}

		fp2, err := os.Create("blue.float")
		if err != nil {
			log.Printf("ERR: %v", err)
			return
		}
		err = binary.Write(fp2, binary.LittleEndian, buf2)
		if err != nil {
			log.Printf("ERR: %v", err)
			return
		}
		fp2.Close()
	*/
	data := make(map[string]interface{})
	var dataMutex sync.Mutex

	data["Package"] = "bluenoisedither"
	data["TileSize"] = size

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		buf := make([]float64, size*size)
		White1D(size, buf)
		Blue1D(size, 1000000, buf)
		Time1D := make([]uint64, size*size)
		for i := range buf {
			Time1D[i] = Float64ToScramble(buf[i])
		}

		dataMutex.Lock()
		data["Time1D"] = Time1D
		dataMutex.Unlock()
		wg.Done()
	}()

	go func() {
		buf := make([]float64, size*size*2)

		White2D(size, buf)
		Blue2D(size, 1000000, buf)

		Lens2D := make([][2]uint64, size*size)

		for i := range Lens2D {
			Lens2D[i][0] = Float64ToScramble(buf[i*2+0])
			Lens2D[i][1] = Float64ToScramble(buf[i*2+1])
		}

		dataMutex.Lock()
		data["Lens2D"] = Lens2D
		dataMutex.Unlock()
		wg.Done()
	}()

	wg.Wait()

	tpl.Execute(os.Stdout, data)
}
