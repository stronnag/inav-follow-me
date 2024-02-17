package geo

import (
	"math"
)

func nm2r(nm float64) float64 {
	return (math.Pi / (180.0 * 60.0)) * nm
}

func r2nm(r float64) float64 {
	return ((180.0 * 60.0) / math.Pi) * r
}

func to_radians(d float32) float64 {
	return float64(d * (math.Pi / 180.0))
}

func to_degrees(r float64) float64 {
	return r * (180.0 / math.Pi)
}

func Csedist(_lat1, _lon1, _lat2, _lon2 float32) (float32, float32) {
	lat1 := to_radians(_lat1)
	lon1 := to_radians(_lon1)
	lat2 := to_radians(_lat2)
	lon2 := to_radians(_lon2)

	p1 := math.Sin((lat1 - lat2) / 2.0)
	p2 := math.Cos(lat1) * math.Cos(lat2)
	p3 := math.Sin((lon2 - lon1) / 2.0)
	d := 2.0 * math.Asin(math.Sqrt((p1*p1)+p2*(p3*p3)))
	d = r2nm(d)
	cse := math.Mod((math.Atan2(math.Sin(lon2-lon1)*math.Cos(lat2),
		math.Cos(lat1)*math.Sin(lat2)-math.Sin(lat1)*math.Cos(lat2)*math.Cos(lon2-lon1))),
		(2.0 * math.Pi))
	cse = to_degrees(cse)
	if cse < 0.0 {
		cse += 360
	}
	return float32(cse), float32(d * 1852.0)
}
