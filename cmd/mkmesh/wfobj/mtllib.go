package wfobj

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
	//"strings"
	//"log"
	"fmt"
	"io"
	"math"
	"unicode/utf8"
)

type Shader interface {
	Write(w io.Writer, prefix string) error
}

type floatparam struct {
	F float32
}

func (m *floatparam) Write(w io.Writer) error {
	_, err := fmt.Fprintf(w, "float %v", m.F)
	return err
}

type rgbparam struct {
	RGB [3]float32
}

func (m *rgbparam) Write(w io.Writer) error {
	_, err := fmt.Fprintf(w, "rgb %v %v %v", m.RGB[0], m.RGB[1], m.RGB[2])
	return err
}

type texparam struct {
	Filename string
}

func (m *texparam) Write(w io.Writer) error {
	_, err := fmt.Fprintf(w, "rgbtex \"%v\"", m.Filename)
	return err
}

type Param interface {
	Write(w io.Writer) error
}

type material struct {
	name   string
	params map[string]Param
}

func (m *material) Write(w io.Writer, prefix string) error {
	_, err := fmt.Fprintf(w, "ShaderStd {\n  Name \"%v:%v\"\n", prefix, m.name)

	if err != nil {
		return err
	}

	for k, v := range m.params {
		fmt.Fprintf(w, "  %v ", k)
		v.Write(w)
		fmt.Fprint(w, "\n")
	}
	fmt.Fprint(w, "\n}\n")
	return nil
}

type lineScanner struct {
	line []byte
	pos  int
}

func (l *lineScanner) next() rune {
	if len(l.line) == 0 {
		return 0
	}

	c, size := utf8.DecodeRune(l.line)
	l.line = l.line[size:]
	if c == utf8.RuneError && size == 1 {
		//log.Print("invalid utf8")
		return l.next()
	}
	return c
}
func (l *lineScanner) init(line string) {
	l.line = []byte(line)
}

func (l *lineScanner) Rest() string {
	//log.Printf("rest: %v", string(l.line))
	return string(l.line)
}

func (l *lineScanner) Token() string {
	var buf bytes.Buffer

	// skip whitespace
L:
	for {
		switch r := l.next(); r {
		case 0, '#':
			return ""

		case ' ', '\t':
			// do nothing, skip
		default:
			buf.WriteRune(r)
			break L
		}
	}

L2:
	for {
		switch r := l.next(); r {
		case 0, ' ', '\t':
			break L2
		default:
			buf.WriteRune(r)
		}
	}
	//log.Printf("tok: %v", buf.String())
	return buf.String()
}

func ParseMtlLib(filename string) (shaders []Shader, err error) {
	fin, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fin.Close()

	//var mtlid int
	scanner := bufio.NewScanner(fin)
	// bytes := make([]byte, DefaultBufferSize)

	var mtl *material

	for scanner.Scan() {
		line := scanner.Text()

		lscan := lineScanner{}
		lscan.init(line)

		cmd := lscan.Token()

		switch cmd {
		case "newmtl":
			name := lscan.Token()

			mtl = &material{name: name, params: map[string]Param{}}
			shaders = append(shaders, mtl)

		case "Ke":

			r, err := strconv.ParseFloat(lscan.Token(), 32)
			g, err := strconv.ParseFloat(lscan.Token(), 32)
			b, err := strconv.ParseFloat(lscan.Token(), 32)

			if err != nil {
				return shaders, err
			}

			if r == 0.0 && g == 0.0 && b == 0.0 {
				continue
			}

			strength := math.Sqrt(r*r + g*g + b*b)
			r /= strength
			g /= strength
			b /= strength
			mtl.params["EmissionColour"] = &rgbparam{[3]float32{float32(r), float32(g), float32(b)}}
			mtl.params["EmissionStrength"] = &floatparam{float32(strength)}

		case "Kd":

			r, err := strconv.ParseFloat(lscan.Token(), 32)
			g, err := strconv.ParseFloat(lscan.Token(), 32)
			b, err := strconv.ParseFloat(lscan.Token(), 32)

			if err != nil {
				return shaders, err
			}
			mtl.params["DiffuseColour"] = &rgbparam{[3]float32{float32(r), float32(g), float32(b)}}

		case "map_Kd":

			mtl.params["DiffuseColour"] = &texparam{lscan.Rest()}

		case "map_bump":
			i := 1
			scale := float32(1.0)
			rest := lscan.Rest()
			if lscan.Token() == "-bm" {
				i++
				scale64, err := strconv.ParseFloat(lscan.Token(), 32)
				scale = float32(scale64)

				if err != nil {
					return shaders, err
				}
				i++
				mtl.params["BumpMap"] = &texparam{lscan.Rest()}
				mtl.params["BumpMapScale"] = &floatparam{scale}
			} else {

				mtl.params["BumpMap"] = &texparam{rest}
				mtl.params["BumpMapScale"] = &floatparam{scale}

			}

		}

	}

	return
}
