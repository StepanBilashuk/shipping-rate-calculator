package shipping

import (
	"errors"
	"fmt"
	"math"
)

type WeightTier struct {
	MaxKg float64
	Rate  Money
}

type WeightRule struct {
	tiers []WeightTier
}

func DefaultWeightTiers() []WeightTier {
	return []WeightTier{
		{MaxKg: 1, Rate: Euros(5)},
		{MaxKg: 5, Rate: Euros(10)},
		{MaxKg: math.Inf(1), Rate: Euros(20)},
	}
}

func NewWeightRule(tiers ...WeightTier) (*WeightRule, error) {
	if len(tiers) == 0 {
		return nil, errors.New("weight tiers: table is empty")
	}
	previous := 0.0
	for i, t := range tiers {
		if math.IsNaN(t.MaxKg) || t.MaxKg <= previous {
			return nil, fmt.Errorf("weight tiers: bound %d (%s kg) must be greater than the previous bound (%s kg)",
				i, formatDecimal(t.MaxKg), formatDecimal(previous))
		}
		previous = t.MaxKg
	}
	if last := tiers[len(tiers)-1].MaxKg; !math.IsInf(last, 1) {
		return nil, fmt.Errorf("weight tiers: the last tier must be unbounded so that every weight is priced, got an upper bound of %s kg",
			formatDecimal(last))
	}

	return &WeightRule{tiers: append([]WeightTier(nil), tiers...)}, nil
}

func (r *WeightRule) Name() string { return "weight_tier" }

func (r *WeightRule) Apply(p Parcel) (LineItem, error) {
	for i, t := range r.tiers {
		if p.WeightKg <= t.MaxKg {
			lower := 0.0
			if i > 0 {
				lower = r.tiers[i-1].MaxKg
			}
			return LineItem{
				Rule:        r.Name(),
				Description: fmt.Sprintf("weight tier (%s)", describeRange(lower, t.MaxKg)),
				Amount:      t.Rate,
			}, nil
		}
	}

	return LineItem{}, fmt.Errorf("weight tier: no tier matches %s kg", formatDecimal(p.WeightKg))
}

func describeRange(lower, upper float64) string {
	switch {
	case math.IsInf(upper, 1):
		return fmt.Sprintf("over %s kg", formatDecimal(lower))
	case lower == 0:
		return fmt.Sprintf("up to %s kg", formatDecimal(upper))
	default:
		return fmt.Sprintf("over %s kg up to %s kg", formatDecimal(lower), formatDecimal(upper))
	}
}
