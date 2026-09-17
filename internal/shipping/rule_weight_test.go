package shipping

import (
	"math"
	"testing"
)

func TestWeightRuleApply(t *testing.T) {
	rule, err := NewWeightRule(DefaultWeightTiers()...)
	if err != nil {
		t.Fatalf("NewWeightRule() error = %v", err)
	}

	tests := []struct {
		name            string
		weightKg        float64
		want            Money
		wantDescription string
	}{
		{"far below the first bound", 0.001, Euros(5), "weight tier (up to 1 kg)"},
		{"just below the first bound", 0.999, Euros(5), "weight tier (up to 1 kg)"},
		{"exactly 1 kg stays in the first tier", 1, Euros(5), "weight tier (up to 1 kg)"},
		{"just above 1 kg moves up", 1.001, Euros(10), "weight tier (over 1 kg up to 5 kg)"},
		{"specification example", 3, Euros(10), "weight tier (over 1 kg up to 5 kg)"},
		{"exactly 5 kg stays in the second tier", 5, Euros(10), "weight tier (over 1 kg up to 5 kg)"},
		{"just above 5 kg moves up", 5.001, Euros(20), "weight tier (over 5 kg)"},
		{"heavy parcel", 1000, Euros(20), "weight tier (over 5 kg)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := rule.Apply(Parcel{WeightKg: tt.weightKg, Zone: ZoneDomestic})
			if err != nil {
				t.Fatalf("Apply(%v kg) error = %v", tt.weightKg, err)
			}
			if item.Amount != tt.want {
				t.Errorf("Apply(%v kg) amount = %s, want %s", tt.weightKg, item.Amount, tt.want)
			}
			if item.Description != tt.wantDescription {
				t.Errorf("Apply(%v kg) description = %q, want %q", tt.weightKg, item.Description, tt.wantDescription)
			}
			if item.Rule != rule.Name() {
				t.Errorf("Apply(%v kg) rule = %q, want %q", tt.weightKg, item.Rule, rule.Name())
			}
		})
	}
}

func TestNewWeightRuleRejectsBadTables(t *testing.T) {
	tests := []struct {
		name  string
		tiers []WeightTier
	}{
		{"empty table", nil},
		{"no open-ended tier", []WeightTier{{MaxKg: 1, Rate: Euros(5)}}},
		{"bounds out of order", []WeightTier{
			{MaxKg: 5, Rate: Euros(10)},
			{MaxKg: 1, Rate: Euros(5)},
			{MaxKg: math.Inf(1), Rate: Euros(20)},
		}},
		{"duplicate bounds", []WeightTier{
			{MaxKg: 1, Rate: Euros(5)},
			{MaxKg: 1, Rate: Euros(10)},
			{MaxKg: math.Inf(1), Rate: Euros(20)},
		}},
		{"non-positive bound", []WeightTier{
			{MaxKg: 0, Rate: Euros(5)},
			{MaxKg: math.Inf(1), Rate: Euros(20)},
		}},
		{"NaN bound", []WeightTier{
			{MaxKg: math.NaN(), Rate: Euros(5)},
			{MaxKg: math.Inf(1), Rate: Euros(20)},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewWeightRule(tt.tiers...); err == nil {
				t.Error("NewWeightRule() = nil error, want an error")
			}
		})
	}
}

func TestWeightRuleSupportsACustomTariff(t *testing.T) {
	rule, err := NewWeightRule(
		WeightTier{MaxKg: 0.5, Rate: Cents(199)},
		WeightTier{MaxKg: math.Inf(1), Rate: Cents(999)},
	)
	if err != nil {
		t.Fatalf("NewWeightRule() error = %v", err)
	}

	for _, tc := range []struct {
		weightKg float64
		want     Money
	}{{0.5, Cents(199)}, {0.51, Cents(999)}} {
		item, err := rule.Apply(Parcel{WeightKg: tc.weightKg, Zone: ZoneDomestic})
		if err != nil {
			t.Fatalf("Apply(%v kg) error = %v", tc.weightKg, err)
		}
		if item.Amount != tc.want {
			t.Errorf("Apply(%v kg) = %s, want %s", tc.weightKg, item.Amount, tc.want)
		}
	}
}

func TestNewWeightRuleCopiesTheTable(t *testing.T) {
	tiers := DefaultWeightTiers()
	rule, err := NewWeightRule(tiers...)
	if err != nil {
		t.Fatalf("NewWeightRule() error = %v", err)
	}

	tiers[0].Rate = Euros(500)

	item, err := rule.Apply(Parcel{WeightKg: 1, Zone: ZoneDomestic})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if item.Amount != Euros(5) {
		t.Errorf("Apply() = %s after the caller changed its slice, want 5.00", item.Amount)
	}
}

func TestWeightRuleWithoutTiersFails(t *testing.T) {
	var rule WeightRule

	if _, err := rule.Apply(Parcel{WeightKg: 3, Zone: ZoneEU}); err == nil {
		t.Error("Apply() = nil error, want an error")
	}
}
