package vgeo

import geoimpl "github.com/imajinyun/knifer-go/internal/geo"

// Coord represents a longitude/latitude coordinate in degrees.
type Coord = geoimpl.Coord

// CoordType identifies a coordinate system.
type CoordType = geoimpl.CoordType

const (
	// WGS84 is the global GPS coordinate system.
	WGS84 CoordType = geoimpl.WGS84
	// GCJ02 is the encrypted Mars coordinate system used by many China map providers.
	GCJ02 CoordType = geoimpl.GCJ02
	// BD09 is Baidu's longitude/latitude coordinate system.
	BD09 CoordType = geoimpl.BD09
	// BD09MC is Baidu's Mercator coordinate system.
	BD09MC CoordType = geoimpl.BD09MC
)

// Error represents an error produced by geo helpers.
type Error = geoimpl.Error

// Convert converts coord between supported coordinate systems.
func Convert(coord Coord, fromType, toType CoordType) (Coord, error) {
	return geoimpl.Convert(coord, fromType, toType)
}

// WGS84ToGCJ02 converts GPS coordinates to GCJ-02.
func WGS84ToGCJ02(coord Coord) Coord { return geoimpl.WGS84ToGCJ02(coord) }

// GCJ02ToWGS84 converts GCJ-02 coordinates back to WGS-84.
func GCJ02ToWGS84(coord Coord) Coord { return geoimpl.GCJ02ToWGS84(coord) }

// GCJ02ToBD09 converts GCJ-02 coordinates to BD-09.
func GCJ02ToBD09(coord Coord) Coord { return geoimpl.GCJ02ToBD09(coord) }

// BD09ToGCJ02 converts BD-09 coordinates to GCJ-02.
func BD09ToGCJ02(coord Coord) Coord { return geoimpl.BD09ToGCJ02(coord) }

// WGS84ToBD09 converts WGS-84 coordinates to BD-09.
func WGS84ToBD09(coord Coord) Coord { return geoimpl.WGS84ToBD09(coord) }

// BD09ToWGS84 converts BD-09 coordinates to WGS-84.
func BD09ToWGS84(coord Coord) Coord { return geoimpl.BD09ToWGS84(coord) }

// Distance returns the great-circle distance between two coordinates in meters.
func Distance(a, b Coord) float64 { return geoimpl.Distance(a, b) }

// InChina reports whether a longitude/latitude pair falls inside a rough mainland China bounding box.
func InChina(lng, lat float64) bool { return geoimpl.InChina(lng, lat) }
