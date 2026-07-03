package geo

import "math"

const (
	xPi          = math.Pi * 3000.0 / 180.0
	earthA       = 6378245.0
	earthEE      = 0.00669342162296594323
	earthRadiusM = 6371000.0
)

var (
	bdMercatorBands = []float64{12890594.86, 8362377.87, 5591021.0, 3481989.83, 1678043.12, 0}
	bdLatLngBands   = []float64{75, 60, 45, 30, 15, 0}
	bdMercatorToLL  = [][]float64{
		{1.410526172116255e-8, 0.00000898305509648872, -1.9939833816331, 200.9824383106796, -187.2403703815547, 91.6087516669843, -23.38765649603339, 2.57121317296198, -0.03801003308653, 17337981.2},
		{-7.435856389565537e-9, 0.000008983055097726239, -0.78625201886289, 96.32687599759846, -1.85204757529826, -59.36935905485877, 47.40033549296737, -16.50741931063887, 2.28786674699375, 10260144.86},
		{-3.030883460898826e-8, 0.00000898305509983578, 0.30071316287616, 59.74293618442277, 7.357984074871, -25.38371002664745, 13.45380521110908, -3.29883767235584, 0.32710905363475, 6856817.37},
		{-1.981981304930552e-8, 0.000008983055099779535, 0.03278182852591, 40.31678527705744, 0.65659298677277, -4.44255534477492, 0.85341911805263, 0.12923347998204, -0.04625736007561, 4482777.06},
		{3.09191371068437e-9, 0.000008983055096812155, 0.00006995724062, 23.10934304144901, -0.00023663490511, -0.6321817810242, -0.00663494467273, 0.03430082397953, -0.00466043876332, 2555164.4},
		{2.890871144776878e-9, 0.000008983055095805407, -3.068298e-8, 7.47137025468032, -0.00000353937994, -0.02145144861037, -0.00001234426596, 0.00010322952773, -0.00000323890364, 826088.5},
	}
	bdLLToMercator = [][]float64{
		{-0.0015702102444, 111320.7020616939, 1704480524535203, -10338987376042340, 26112667856603880, -35149669176653700, 26595700718403920, -10725012454188240, 1800819912950474, 82.5},
		{0.0008277824516172526, 111320.7020463578, 647795574.6671607, -4082003173.641316, 10774905663.51142, -15171875531.51559, 12053065338.62167, -5124939663.577472, 913311935.9512032, 67.5},
		{0.00337398766765, 111320.7020202162, 4481351.045890365, -23393751.19931662, 79682215.47186455, -115964993.2797253, 97236711.15602145, -43661946.33752821, 8477230.501135234, 52.5},
		{0.00220636496208, 111320.7020209128, 51751.86112841131, 3796837.749470245, 992013.7397791013, -1221952.21711287, 1340652.697009075, -620943.6990984312, 144416.9293806241, 37.5},
		{-0.0003441963504368392, 111320.7020576856, 278.2353980772752, 2485758.690035394, 6070.750963243378, 54821.18345352118, 9540.606633304236, -2710.55326746645, 1405.483844121726, 22.5},
		{-0.0003218135878613132, 111320.7020701615, 0.00369383431289, 823725.6402795718, 0.46104986909093, 2351.343141331292, 1.58060784298199, 8.77738589078284, 0.37238884252424, 7.45},
	}
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
	if !isSupportedCoordType(fromType) {
		return Coord{}, invalidInputf("unsupported coordinate type %s", fromType)
	}
	if !isSupportedCoordType(toType) {
		return Coord{}, invalidInputf("unsupported coordinate type %s", toType)
	}
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
		case BD09MC:
			return bd09ToBD09MC(WGS84ToBD09(coord)), nil
		}
	case GCJ02:
		switch toType {
		case WGS84:
			return GCJ02ToWGS84(coord), nil
		case BD09:
			return GCJ02ToBD09(coord), nil
		case BD09MC:
			return bd09ToBD09MC(GCJ02ToBD09(coord)), nil
		}
	case BD09:
		switch toType {
		case GCJ02:
			return BD09ToGCJ02(coord), nil
		case WGS84:
			return BD09ToWGS84(coord), nil
		case BD09MC:
			return bd09ToBD09MC(coord), nil
		}
	case BD09MC:
		bd := bd09MCToBD09(coord)
		switch toType {
		case BD09:
			return bd, nil
		case GCJ02:
			return BD09ToGCJ02(bd), nil
		case WGS84:
			return BD09ToWGS84(bd), nil
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
	if h > 1 {
		h = 1
	}
	return 2 * earthRadiusM * math.Asin(math.Sqrt(h))
}

// InChina reports whether a longitude/latitude pair falls inside a rough mainland China bounding box.
func InChina(lng, lat float64) bool {
	return lng >= 72.004 && lng <= 137.8347 && lat >= 0.8293 && lat <= 55.8271
}

func isSupportedCoordType(coordType CoordType) bool {
	switch coordType {
	case WGS84, GCJ02, BD09, BD09MC:
		return true
	default:
		return false
	}
}

func bd09ToBD09MC(coord Coord) Coord {
	lng := loop(coord.Lng, -180, 180)
	lat := clamp(coord.Lat, -74, 74)
	var factors []float64
	for i, band := range bdLatLngBands {
		if lat >= band {
			factors = bdLLToMercator[i]
			break
		}
	}
	if factors == nil {
		for i, band := range bdLatLngBands {
			if lat <= -band {
				factors = bdLLToMercator[i]
				break
			}
		}
	}
	if factors == nil {
		factors = bdLLToMercator[len(bdLLToMercator)-1]
	}
	return bdConvertor(Coord{Lng: lng, Lat: lat}, factors)
}

func bd09MCToBD09(coord Coord) Coord {
	absLat := math.Abs(coord.Lat)
	var factors []float64
	for i, band := range bdMercatorBands {
		if absLat >= band {
			factors = bdMercatorToLL[i]
			break
		}
	}
	if factors == nil {
		factors = bdMercatorToLL[len(bdMercatorToLL)-1]
	}
	return bdConvertor(coord, factors)
}

func bdConvertor(coord Coord, factors []float64) Coord {
	x := factors[0] + factors[1]*math.Abs(coord.Lng)
	y := math.Abs(coord.Lat) / factors[9]
	y2 := y * y
	y3 := y2 * y
	y4 := y3 * y
	y5 := y4 * y
	y6 := y5 * y
	y = factors[2] + factors[3]*y + factors[4]*y2 + factors[5]*y3 +
		factors[6]*y4 + factors[7]*y5 + factors[8]*y6
	x *= sign(coord.Lng)
	y *= sign(coord.Lat)
	return Coord{Lng: x, Lat: y}
}

func loop(value, min, max float64) float64 {
	for value > max {
		value -= max - min
	}
	for value < min {
		value += max - min
	}
	return value
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func sign(value float64) float64 {
	if value < 0 {
		return -1
	}
	return 1
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
