package shipping

import (
	"math"
	"strings"
	"testing"
)

func validParcel() Parcel {
	return Parcel{WeightKg: 3, LengthCm: 20, WidthCm: 10, HeightCm: 5, Zone: ZoneEU}
}

func TestParcelValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Parcel)
		wantErr string
	}{
		{"valid parcel", func(*Parcel) {}, ""},
		{"tiny but positive values are valid", func(p *Parcel) { p.WeightKg = 0.001 }, ""},
		{"zero weight", func(p *Parcel) { p.WeightKg = 0 }, "weight"},
		{"negative weight", func(p *Parcel) { p.WeightKg = -1 }, "weight"},
		{"NaN weight", func(p *Parcel) { p.WeightKg = math.NaN() }, "weight"},
		{"infinite weight", func(p *Parcel) { p.WeightKg = math.Inf(1) }, "weight"},
		{"zero length", func(p *Parcel) { p.LengthCm = 0 }, "length"},
		{"negative width", func(p *Parcel) { p.WidthCm = -2 }, "width"},
		{"zero height", func(p *Parcel) { p.HeightCm = 0 }, "height"},
		{"missing zone", func(p *Parcel) { p.Zone = "" }, "zone: required"},
		{"unknown zone", func(p *Parcel) { p.Zone = "moon" }, `zone: "moon" is not one of`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validParcel()
			tt.mutate(&p)

			err := p.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want an error mentioning %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() = %q, want it to mention %q", err, tt.wantErr)
			}
		})
	}
}

func TestParcelValidateReportsAllProblems(t *testing.T) {
	p := Parcel{WeightKg: 0, LengthCm: -1, WidthCm: 0, HeightCm: 0, Zone: "moon"}

	err := p.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want errors")
	}
	for _, field := range []string{"weight", "length", "width", "height", "zone"} {
		if !strings.Contains(err.Error(), field) {
			t.Errorf("Validate() = %q, want it to mention %q", err, field)
		}
	}
}

func TestParcelString(t *testing.T) {
	p := validParcel()
	p.WeightKg = 2.5

	if got, want := p.String(), "2.5 kg, 20x10x5 cm, zone eu"; got != want {
		t.Errorf("Parcel.String() = %q, want %q", got, want)
	}
}
