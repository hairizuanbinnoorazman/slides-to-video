package pdfslideimages

import (
	"testing"
)

func TestSetIdemKey(t *testing.T) {
	type args struct {
		idemKey string
	}
	tests := []struct {
		name string
		args args
		want func(*PDFSlideImages)
	}{
		{
			name: "simple",
			args: args{
				idemKey: "accaca",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := SetIdemKey(tt.args.idemKey)
			z := PDFSlideImages{}
			a(&z)
			if z.IdemKey != tt.args.idemKey {
				t.Errorf("expected: %v actual: %v", tt.args.idemKey, z.IdemKey)
			}
		})
	}
}
