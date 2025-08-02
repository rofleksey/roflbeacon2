package database

import "roflbeacon2/pkg/util"

func (f *Fence) Contains(lat float64, lon float64, accuracy float64) bool {
	distMeters := util.HaversineDistance(f.Latitude, f.Longitude, lat, lon)

	return distMeters < f.Radius+accuracy
}
