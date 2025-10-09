package restutil

const unitsPerMinute = 100

// UnitsPerMinute exposes the scaling factor used by the tracker API for rest time.
const UnitsPerMinute = unitsPerMinute

// MinutesFromUnits converts the raw rest units returned by the API into minutes.
func MinutesFromUnits(units int) float64 {
	return float64(units) / UnitsPerMinute
}

// UnitsFromMinutes converts minutes into the raw units expected by the API.
func UnitsFromMinutes(minutes int) int {
	return minutes * UnitsPerMinute
}
