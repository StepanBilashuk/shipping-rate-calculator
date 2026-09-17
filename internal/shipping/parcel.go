package shipping

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

type Parcel struct {
	WeightKg float64 `json:"weight"`
	LengthCm float64 `json:"length"`
	WidthCm  float64 `json:"width"`
	HeightCm float64 `json:"height"`
	Zone     Zone    `json:"zone"`
}

func (p Parcel) Validate() error {
	return errors.Join(
		checkPositive("weight", p.WeightKg),
		checkPositive("length", p.LengthCm),
		checkPositive("width", p.WidthCm),
		checkPositive("height", p.HeightCm),
		p.checkZone(),
	)
}

func (p Parcel) checkZone() error {
	switch {
	case p.Zone == "":
		return fmt.Errorf("zone: required, one of %s", joinZones(Zones()))
	case !p.Zone.Valid():
		return fmt.Errorf("zone: %q is not one of %s", p.Zone, joinZones(Zones()))
	default:
		return nil
	}
}

func checkPositive(field string, v float64) error {
	switch {
	case math.IsNaN(v):
		return fmt.Errorf("%s: must be a number", field)
	case math.IsInf(v, 0):
		return fmt.Errorf("%s: must be finite", field)
	case v <= 0:
		return fmt.Errorf("%s: must be greater than 0, got %s", field, formatDecimal(v))
	default:
		return nil
	}
}

func (p Parcel) String() string {
	return fmt.Sprintf("%s kg, %sx%sx%s cm, zone %s",
		formatDecimal(p.WeightKg),
		formatDecimal(p.LengthCm), formatDecimal(p.WidthCm), formatDecimal(p.HeightCm),
		p.Zone)
}

func formatDecimal(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
