package geo

import "math"

const (
	xPi          = math.Pi * 3000.0 / 180.0
	earthA       = 6378245.0
	earthEE      = 0.00669342162296594323
	earthRadiusM = 6371000.0
)

// Coord represents a longitude/latitude coordinate in degrees.
type Coord struct {
	Lng float64
	Lat float64
}

// CoordType identifies a coordinate system.
type CoordType string

const (
	// WGS84 is the global GPS coordinate system.
	WGS84 CoordType = "WGS84"
	// GCJ02 is the encrypted Mars coordinate system used by many China map providers.
	GCJ02 CoordType = "GCJ02"
	// BD09 is Baidu's longitude/latitude coordinate system.
	BD09 CoordType = "BD09"
	// BD09MC is Baidu's Mercator coordinate system.
	BD09MC CoordType = "BD09MC"
)

// Convert converts coord between supported coordinate systems.
func Convert(coord Coord, fromType, toType CoordType) (Coord, error) {
	if fromType == toType {
		return coord, nil
	}

	switch fromType {
	case WGS84:
		switch toType {
		case GCJ02:
			return WGS84ToGCJ02(coord), nil
		case BD09:
			return WGS84ToBD09(coord), nil
		}
	case GCJ02:
		switch toType {
		case WGS84:
			return GCJ02ToWGS84(coord), nil
		case BD09:
			return GCJ02ToBD09(coord), nil
		}
	case BD09:
		switch toType {
		case GCJ02:
			return BD09ToGCJ02(coord), nil
		case WGS84:
			return BD09ToWGS84(coord), nil
		}
	}

	return Coord{}, invalidInputf("unsupported coordinate conversion %s to %s", fromType, toType)
}

// WGS84ToGCJ02 converts GPS coordinates to GCJ-02.
func WGS84ToGCJ02(coord Coord) Coord {
	if !InChina(coord.Lng, coord.Lat) {
		return coord
	}
	dLat := transformLat(coord.Lng-105.0, coord.Lat-35.0)
	dLng := transformLng(coord.Lng-105.0, coord.Lat-35.0)
	radLat := coord.Lat / 180.0 * math.Pi
	magic := math.Sin(radLat)
	magic = 1 - earthEE*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dLat = (dLat * 180.0) / ((earthA * (1 - earthEE)) / (magic * sqrtMagic) * math.Pi)
	dLng = (dLng * 180.0) / (earthA / sqrtMagic * math.Cos(radLat) * math.Pi)
	return Coord{Lng: coord.Lng + dLng, Lat: coord.Lat + dLat}
}

// GCJ02ToWGS84 converts GCJ-02 coordinates back to WGS-84.
func GCJ02ToWGS84(coord Coord) Coord {
	if !InChina(coord.Lng, coord.Lat) {
		return coord
	}
	gcj := WGS84ToGCJ02(coord)
	return Coord{Lng: coord.Lng*2 - gcj.Lng, Lat: coord.Lat*2 - gcj.Lat}
}

// GCJ02ToBD09 converts GCJ-02 coordinates to BD-09.
func GCJ02ToBD09(coord Coord) Coord {
	z := math.Sqrt(coord.Lng*coord.Lng+coord.Lat*coord.Lat) + 0.00002*math.Sin(coord.Lat*xPi)
	theta := math.Atan2(coord.Lat, coord.Lng) + 0.000003*math.Cos(coord.Lng*xPi)
	return Coord{Lng: z*math.Cos(theta) + 0.0065, Lat: z*math.Sin(theta) + 0.006}
}

// BD09ToGCJ02 converts BD-09 coordinates to GCJ-02.
func BD09ToGCJ02(coord Coord) Coord {
	x := coord.Lng - 0.0065
	y := coord.Lat - 0.006
	z := math.Sqrt(x*x+y*y) - 0.00002*math.Sin(y*xPi)
	theta := math.Atan2(y, x) - 0.000003*math.Cos(x*xPi)
	return Coord{Lng: z * math.Cos(theta), Lat: z * math.Sin(theta)}
}

// WGS84ToBD09 converts WGS-84 coordinates to BD-09.
func WGS84ToBD09(coord Coord) Coord {
	return GCJ02ToBD09(WGS84ToGCJ02(coord))
}

// BD09ToWGS84 converts BD-09 coordinates to WGS-84.
func BD09ToWGS84(coord Coord) Coord {
	return GCJ02ToWGS84(BD09ToGCJ02(coord))
}

// Distance returns the great-circle distance between two coordinates in meters.
func Distance(a, b Coord) float64 {
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLng := (b.Lng - a.Lng) * math.Pi / 180

	sinLat := math.Sin(dLat / 2)
	sinLng := math.Sin(dLng / 2)
	h := sinLat*sinLat + math.Cos(lat1)*math.Cos(lat2)*sinLng*sinLng
	return 2 * earthRadiusM * math.Asin(math.Sqrt(h))
}

// InChina reports whether a longitude/latitude pair falls inside a rough mainland China bounding box.
func InChina(lng, lat float64) bool {
	return lng >= 72.004 && lng <= 137.8347 && lat >= 0.8293 && lat <= 55.8271
}

func transformLat(x, y float64) float64 {
	ret := -100.0 + 2.0*x + 3.0*y + 0.2*y*y + 0.1*x*y + 0.2*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(y*math.Pi) + 40.0*math.Sin(y/3.0*math.Pi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(y/12.0*math.Pi) + 320*math.Sin(y*math.Pi/30.0)) * 2.0 / 3.0
	return ret
}

func transformLng(x, y float64) float64 {
	ret := 300.0 + x + 2.0*y + 0.1*x*x + 0.1*x*y + 0.1*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(x*math.Pi) + 40.0*math.Sin(x/3.0*math.Pi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(x/12.0*math.Pi) + 300.0*math.Sin(x/30.0*math.Pi)) * 2.0 / 3.0
	return ret
}
