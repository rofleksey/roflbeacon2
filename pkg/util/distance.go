package util

import (
	"github.com/LucaTheHacker/go-haversine"
)

func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	hs := haversine.Distance(haversine.NewCoordinates(lat1, lon1), haversine.NewCoordinates(lat2, lon2))
	distKm := hs.Kilometers()
	return distKm * 1000
}
