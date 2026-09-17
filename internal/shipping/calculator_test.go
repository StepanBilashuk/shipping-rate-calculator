package shipping

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func TestCalculatorQuoteTotals(t *testing.T) {
	calc, err := NewDefaultCalculator()
	if err != nil {
		t.Fatalf("NewDefaultCalculator() error = %v", err)
	}

	tests := []struct {
		name     string
		weightKg float64
		zone     Zone
		want     string
	}{
		{"1 kg domestic", 1, ZoneDomestic, "5.00"},
		{"1 kg eu", 1, ZoneEU, "13.00"},
		{"1 kg international", 1, ZoneInternational, "20.00"},
		{"5 kg domestic", 5, ZoneDomestic, "10.00"},
		{"5 kg eu", 5, ZoneEU, "18.00"},
		{"5 kg international", 5, ZoneInternational, "25.00"},
		{"6 kg domestic", 6, ZoneDomestic, "20.00"},
		{"6 kg eu", 6, ZoneEU, "28.00"},
		{"6 kg international", 6, ZoneInternational, "35.00"},
		{"specification example: 3 kg to the EU", 3, ZoneEU, "18.00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := calc.Quote(Parcel{WeightKg: tt.weightKg, LengthCm: 20, WidthCm: 10, HeightCm: 5, Zone: tt.zone})
			if err != nil {
				t.Fatalf("Quote() error = %v", err)
			}
			if got := q.Total.String(); got != tt.want {
				t.Errorf("Quote() total = %s, want %s", got, tt.want)
			}
			if q.Currency != Currency {
				t.Errorf("Quote() currency = %q, want %q", q.Currency, Currency)
			}
		})
	}
}

func TestCalculatorQuoteBreakdown(t *testing.T) {
	calc, err := NewDefaultCalculator()
	if err != nil {
		t.Fatalf("NewDefaultCalculator() error = %v", err)
	}
	parcel := Parcel{WeightKg: 3, LengthCm: 20, WidthCm: 10, HeightCm: 5, Zone: ZoneEU}

	q, err := calc.Quote(parcel)
	if err != nil {
		t.Fatalf("Quote() error = %v", err)
	}

	if q.Parcel != parcel {
		t.Errorf("Quote() parcel = %+v, want %+v", q.Parcel, parcel)
	}
	if len(q.Breakdown) != 2 {
		t.Fatalf("Quote() breakdown has %d items, want 2: %+v", len(q.Breakdown), q.Breakdown)
	}
	if got, want := q.Breakdown[0].Rule, "weight_tier"; got != want {
		t.Errorf("first line item rule = %q, want %q", got, want)
	}
	if got, want := q.Breakdown[0].Amount.String(), "10.00"; got != want {
		t.Errorf("weight tier amount = %s, want %s", got, want)
	}
	if got, want := q.Breakdown[1].Rule, "zone_surcharge"; got != want {
		t.Errorf("second line item rule = %q, want %q", got, want)
	}
	if got, want := q.Breakdown[1].Amount.String(), "8.00"; got != want {
		t.Errorf("zone surcharge amount = %s, want %s", got, want)
	}

	var sum Money
	for _, item := range q.Breakdown {
		sum += item.Amount
	}
	if sum != q.Total {
		t.Errorf("breakdown sums to %s, but the total is %s", sum, q.Total)
	}
}

func TestCalculatorQuoteRejectsInvalidParcels(t *testing.T) {
	calc, err := NewDefaultCalculator()
	if err != nil {
		t.Fatalf("NewDefaultCalculator() error = %v", err)
	}

	tests := []struct {
		name   string
		parcel Parcel
	}{
		{"zero value", Parcel{}},
		{"missing weight", Parcel{LengthCm: 20, WidthCm: 10, HeightCm: 5, Zone: ZoneEU}},
		{"missing dimensions", Parcel{WeightKg: 3, Zone: ZoneEU}},
		{"unknown zone", Parcel{WeightKg: 3, LengthCm: 20, WidthCm: 10, HeightCm: 5, Zone: "moon"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := calc.Quote(tt.parcel)
			if err == nil {
				t.Fatalf("Quote(%+v) = %+v, want an error", tt.parcel, q)
			}
			if q.Total != 0 || q.Breakdown != nil {
				t.Errorf("Quote() returned a partial quote %+v, want the zero value", q)
			}
		})
	}
}

func TestCalculatorAppliesAdditionalRules(t *testing.T) {
	weight, err := NewWeightRule(DefaultWeightTiers()...)
	if err != nil {
		t.Fatalf("NewWeightRule() error = %v", err)
	}
	zone, err := NewZoneRule(DefaultZoneSurcharges())
	if err != nil {
		t.Fatalf("NewZoneRule() error = %v", err)
	}
	handling := stubRule{name: "handling_fee", amount: Cents(250)}

	calc, err := NewCalculator(weight, zone, handling)
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	q, err := calc.Quote(Parcel{WeightKg: 3, LengthCm: 20, WidthCm: 10, HeightCm: 5, Zone: ZoneEU})
	if err != nil {
		t.Fatalf("Quote() error = %v", err)
	}
	if got, want := q.Total.String(), "20.50"; got != want {
		t.Errorf("Quote() total = %s, want %s", got, want)
	}
	if len(q.Breakdown) != 3 {
		t.Errorf("Quote() breakdown has %d items, want 3: %+v", len(q.Breakdown), q.Breakdown)
	}
}

func TestCalculatorPropagatesRuleErrors(t *testing.T) {
	boom := errors.New("tariff service unavailable")
	calc, err := NewCalculator(stubRule{name: "flaky", err: boom})
	if err != nil {
		t.Fatalf("NewCalculator() error = %v", err)
	}

	_, err = calc.Quote(Parcel{WeightKg: 3, LengthCm: 20, WidthCm: 10, HeightCm: 5, Zone: ZoneEU})
	if !errors.Is(err, boom) {
		t.Fatalf("Quote() error = %v, want it to wrap %v", err, boom)
	}
	if !strings.Contains(err.Error(), "flaky") {
		t.Errorf("Quote() error = %q, want it to name the failing rule", err)
	}
}

func TestNewCalculatorRequiresRules(t *testing.T) {
	if _, err := NewCalculator(); err == nil {
		t.Error("NewCalculator() = nil error, want an error")
	}
}

func TestNewCalculatorFromTariff(t *testing.T) {
	tests := []struct {
		name       string
		tiers      []WeightTier
		surcharges map[Zone]Money
		wantErr    string
	}{
		{
			name:       "valid tariff",
			tiers:      DefaultWeightTiers(),
			surcharges: DefaultZoneSurcharges(),
		},
		{
			name:       "invalid weight tiers are rejected",
			tiers:      []WeightTier{{MaxKg: 1, Rate: Euros(5)}},
			surcharges: DefaultZoneSurcharges(),
			wantErr:    "weight tiers",
		},
		{
			name:       "incomplete zone table is rejected",
			tiers:      DefaultWeightTiers(),
			surcharges: map[Zone]Money{ZoneDomestic: Euros(0)},
			wantErr:    "zone surcharges",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc, err := NewCalculatorFromTariff(tt.tiers, tt.surcharges)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("NewCalculatorFromTariff() error = %v, want nil", err)
				}
				if calc == nil {
					t.Fatal("NewCalculatorFromTariff() = nil calculator, want one")
				}
				return
			}
			if err == nil {
				t.Fatalf("NewCalculatorFromTariff() = %v, want an error mentioning %q", calc, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("NewCalculatorFromTariff() error = %q, want it to mention %q", err, tt.wantErr)
			}
		})
	}
}

func TestCalculatorFromCustomTariff(t *testing.T) {
	calc, err := NewCalculatorFromTariff(
		[]WeightTier{{MaxKg: math.Inf(1), Rate: Euros(1)}},
		map[Zone]Money{ZoneDomestic: Euros(2), ZoneEU: Euros(3), ZoneInternational: Euros(4)},
	)
	if err != nil {
		t.Fatalf("NewCalculatorFromTariff() error = %v", err)
	}

	q, err := calc.Quote(Parcel{WeightKg: 900, LengthCm: 1, WidthCm: 1, HeightCm: 1, Zone: ZoneEU})
	if err != nil {
		t.Fatalf("Quote() error = %v", err)
	}
	if got, want := q.Total.String(), "4.00"; got != want {
		t.Errorf("Quote() total = %s, want %s", got, want)
	}
}

type stubRule struct {
	name   string
	amount Money
	err    error
}

func (r stubRule) Name() string { return r.name }

func (r stubRule) Apply(Parcel) (LineItem, error) {
	if r.err != nil {
		return LineItem{}, r.err
	}
	return LineItem{Rule: r.name, Description: r.name, Amount: r.amount}, nil
}
