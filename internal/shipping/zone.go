package shipping

import "strings"

type Zone string

const (
	ZoneDomestic      Zone = "domestic"
	ZoneEU            Zone = "eu"
	ZoneInternational Zone = "international"
)

func Zones() []Zone {
	return []Zone{ZoneDomestic, ZoneEU, ZoneInternational}
}

func (z Zone) Valid() bool {
	for _, known := range Zones() {
		if z == known {
			return true
		}
	}
	return false
}

func (z Zone) String() string { return string(z) }

func joinZones(zones []Zone) string {
	names := make([]string, len(zones))
	for i, z := range zones {
		names[i] = string(z)
	}
	return strings.Join(names, ", ")
}
