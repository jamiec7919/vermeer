package core

import (
	"encoding/binary"
	"os"
	"testing"
	"vermeer/sobol"
)

func TestQMC(t *testing.T) {
	res := 512
	buf := make([]float32, res*res)

	for k := 0; k < 1000000; k++ {
		x := sobol.RadicalInvVdC(uint64(k), 0)
		y := sobol.RadicalInvSobol(uint64(k), 0)

		xi := int(x * float64(res))
		yi := int(y * float64(res))

		buf[xi+yi*res]++
	}

	fp, err := os.Create("out.float")
	defer fp.Close()
	if err != nil {
		t.Logf("error %v", err)
		return
	}
	err = binary.Write(fp, binary.LittleEndian, buf)
	if err != nil {
		return
	}

}
