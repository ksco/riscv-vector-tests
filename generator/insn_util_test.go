package generator

import "testing"

func TestVLEN_Valid(t *testing.T) {
	tests := map[VLEN]bool{
		16:     false,
		32:     true,
		64:     true,
		128:    true,
		129:    false,
		256:    true,
		258:    false,
		65536:  true,
		131072: false,
	}
	for vlen, expected := range tests {
		if vlen.Valid() != expected {
			t.Fatalf("VLEN (%v) test failed, expected %v, got %v",
				vlen, expected, vlen.Valid())
		}
	}
}
