package misc

import (
	"testing"
)

func TestHex2RGBA(t *testing.T) {
	hex1 := "#4B0082"
	c1 := HexColor(hex1)
	t.Log(c1)
	if c1.R != 75 || c1.G != 0 || c1.B != 130 {
		t.Fail()
	}

}

func TestBuffer(t *testing.T) {
	buf := []int{1, 2, 3, 0, 0, 0, 4, 5, 6}

	copy(buf[2:], buf[5:6+3])

	t.Log(buf)
}
