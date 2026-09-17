package shipping

import (
	"fmt"
	"maps"
)

type ZoneRule struct {
	surcharges map[Zone]Money
}

func DefaultZoneSurcharges() map[Zone]Money {
	return map[Zone]Money{
		ZoneDomestic:      Euros(0),
		ZoneEU:            Euros(8),
		ZoneInternational: Euros(15),
	}
}

func NewZoneRule(surcharges map[Zone]Money) (*ZoneRule, error) {
	for _, z := range Zones() {
		if _, ok := surcharges[z]; !ok {
			return nil, fmt.Errorf("zone surcharges: missing entry for zone %q", z)
		}
	}
	for z := range surcharges {
		if !z.Valid() {
			return nil, fmt.Errorf("zone surcharges: unknown zone %q", z)
		}
	}

	return &ZoneRule{surcharges: maps.Clone(surcharges)}, nil
}

func (r *ZoneRule) Name() string { return "zone_surcharge" }

func (r *ZoneRule) Apply(p Parcel) (LineItem, error) {
	amount, ok := r.surcharges[p.Zone]
	if !ok {
		return LineItem{}, fmt.Errorf("zone surcharge: no surcharge configured for zone %q", p.Zone)
	}
	return LineItem{
		Rule:        r.Name(),
		Description: fmt.Sprintf("zone surcharge (%s)", p.Zone),
		Amount:      amount,
	}, nil
}
