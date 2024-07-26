package misc

import "testing"

func TestHex2RGBA(t *testing.T) {
	hex1 := "#4B0082"
	c1 := HexColor(hex1)
	t.Log(c1)
	if c1.R != 75 || c1.G != 0 || c1.B != 130 {
		t.Fail()
	}

}
