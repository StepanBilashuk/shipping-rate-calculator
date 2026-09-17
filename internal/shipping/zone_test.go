package shipping

import "testing"

func TestZoneValid(t *testing.T) {
	tests := []struct {
		name string
		zone Zone
		want bool
	}{
		{"domestic", ZoneDomestic, true},
		{"eu", ZoneEU, true},
		{"international", ZoneInternational, true},
		{"unknown zone", Zone("moon"), false},
		{"near miss is not accepted", Zone("europe"), false},
		{"wrong case is not accepted here", Zone("EU"), false},
		{"zero value", Zone(""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.zone.Valid(); got != tt.want {
				t.Errorf("Zone(%q).Valid() = %v, want %v", tt.zone, got, tt.want)
			}
		})
	}
}

func TestZonesAreValid(t *testing.T) {
	zones := Zones()
	if len(zones) == 0 {
		t.Fatal("Zones() is empty")
	}
	for _, z := range zones {
		if !z.Valid() {
			t.Errorf("Zones() contains %q, which Valid() rejects", z)
		}
	}
}

func TestZonesCannotBeChangedByCallers(t *testing.T) {
	zones := Zones()
	zones[0] = Zone("moon")

	if got := Zones()[0]; got == Zone("moon") {
		t.Error("changing the slice returned by Zones() changed the known zones")
	}
}

func TestZoneString(t *testing.T) {
	if got, want := ZoneInternational.String(), "international"; got != want {
		t.Errorf("Zone.String() = %q, want %q", got, want)
	}
}
