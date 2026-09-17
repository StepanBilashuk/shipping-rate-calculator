package shipping

import "testing"

func TestZoneRuleApply(t *testing.T) {
	rule, err := NewZoneRule(DefaultZoneSurcharges())
	if err != nil {
		t.Fatalf("NewZoneRule() error = %v", err)
	}

	tests := []struct {
		name            string
		zone            Zone
		want            Money
		wantDescription string
	}{
		{"domestic adds nothing", ZoneDomestic, Euros(0), "zone surcharge (domestic)"},
		{"eu", ZoneEU, Euros(8), "zone surcharge (eu)"},
		{"international", ZoneInternational, Euros(15), "zone surcharge (international)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := rule.Apply(Parcel{WeightKg: 1, Zone: tt.zone})
			if err != nil {
				t.Fatalf("Apply(%q) error = %v", tt.zone, err)
			}
			if item.Amount != tt.want {
				t.Errorf("Apply(%q) amount = %s, want %s", tt.zone, item.Amount, tt.want)
			}
			if item.Description != tt.wantDescription {
				t.Errorf("Apply(%q) description = %q, want %q", tt.zone, item.Description, tt.wantDescription)
			}
			if item.Rule != rule.Name() {
				t.Errorf("Apply(%q) rule = %q, want %q", tt.zone, item.Rule, rule.Name())
			}
		})
	}
}

func TestNewZoneRuleRejectsBadTables(t *testing.T) {
	tests := []struct {
		name       string
		surcharges map[Zone]Money
	}{
		{"empty table", map[Zone]Money{}},
		{"missing international", map[Zone]Money{
			ZoneDomestic: Euros(0),
			ZoneEU:       Euros(8),
		}},
		{"unknown zone in the table", map[Zone]Money{
			ZoneDomestic:      Euros(0),
			ZoneEU:            Euros(8),
			ZoneInternational: Euros(15),
			Zone("europe"):    Euros(8),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewZoneRule(tt.surcharges); err == nil {
				t.Error("NewZoneRule() = nil error, want an error")
			}
		})
	}
}

func TestZoneRuleRejectsUnknownZone(t *testing.T) {
	rule, err := NewZoneRule(DefaultZoneSurcharges())
	if err != nil {
		t.Fatalf("NewZoneRule() error = %v", err)
	}

	if _, err := rule.Apply(Parcel{WeightKg: 1, Zone: "moon"}); err == nil {
		t.Error("Apply() = nil error, want an error")
	}
}

func TestNewZoneRuleCopiesTheTable(t *testing.T) {
	surcharges := DefaultZoneSurcharges()
	rule, err := NewZoneRule(surcharges)
	if err != nil {
		t.Fatalf("NewZoneRule() error = %v", err)
	}

	surcharges[ZoneEU] = Euros(500)

	item, err := rule.Apply(Parcel{WeightKg: 1, Zone: ZoneEU})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if item.Amount != Euros(8) {
		t.Errorf("Apply() = %s after the caller changed its map, want 8.00", item.Amount)
	}
}
