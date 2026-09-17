package shipping

import (
	"errors"
	"fmt"
)

const Currency = "EUR"

type Quote struct {
	Parcel    Parcel     `json:"parcel"`
	Breakdown []LineItem `json:"breakdown"`
	Total     Money      `json:"total"`
	Currency  string     `json:"currency"`
}

type Calculator struct {
	rules []Rule
}

func NewCalculator(rules ...Rule) (*Calculator, error) {
	if len(rules) == 0 {
		return nil, errors.New("calculator: at least one rule is required")
	}
	return &Calculator{rules: append([]Rule(nil), rules...)}, nil
}

func NewCalculatorFromTariff(tiers []WeightTier, surcharges map[Zone]Money) (*Calculator, error) {
	weight, err := NewWeightRule(tiers...)
	if err != nil {
		return nil, err
	}
	zone, err := NewZoneRule(surcharges)
	if err != nil {
		return nil, err
	}
	return NewCalculator(weight, zone)
}

func NewDefaultCalculator() (*Calculator, error) {
	return NewCalculatorFromTariff(DefaultWeightTiers(), DefaultZoneSurcharges())
}

func (c *Calculator) Quote(p Parcel) (Quote, error) {
	if err := p.Validate(); err != nil {
		return Quote{}, fmt.Errorf("invalid parcel: %w", err)
	}

	breakdown := make([]LineItem, 0, len(c.rules))
	var total Money
	for _, rule := range c.rules {
		item, err := rule.Apply(p)
		if err != nil {
			return Quote{}, fmt.Errorf("rule %q: %w", rule.Name(), err)
		}
		breakdown = append(breakdown, item)
		total += item.Amount
	}

	return Quote{Parcel: p, Breakdown: breakdown, Total: total, Currency: Currency}, nil
}
