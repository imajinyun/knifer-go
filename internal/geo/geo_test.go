package geo

import (
	"errors"
	"math"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
)

func TestCoordinateConversion(t *testing.T) {
	wgs := Coord{Lng: 116.397389, Lat: 39.908722}
	gcj := WGS84ToGCJ02(wgs)
	wantGCJ := Coord{Lng: 116.403632, Lat: 39.910125}
	if distanceApprox(gcj, wantGCJ) > 1 {
		t.Fatalf("WGS84ToGCJ02 = %#v, want near %#v", gcj, wantGCJ)
	}

	back := GCJ02ToWGS84(gcj)
	if Distance(wgs, back) > 1 {
		t.Fatalf("GCJ02ToWGS84 distance = %.2fm", Distance(wgs, back))
	}

	bd := WGS84ToBD09(wgs)
	roundTrip := BD09ToWGS84(bd)
	if Distance(wgs, roundTrip) > 2 {
		t.Fatalf("BD09 round trip distance = %.2fm", Distance(wgs, roundTrip))
	}
}

func TestConvertBD09MCRoundTrip(t *testing.T) {
	bd := Coord{Lng: 116.410005, Lat: 39.916465}
	mc, err := Convert(bd, BD09, BD09MC)
	if err != nil {
		t.Fatalf("Convert BD09 to BD09MC error = %v", err)
	}
	if mc.Lng < 12958000 || mc.Lng > 12961000 || mc.Lat < 4825000 || mc.Lat > 4827000 {
		t.Fatalf("Convert BD09 to BD09MC = %#v, want Baidu Mercator range near Beijing", mc)
	}

	roundTrip, err := Convert(mc, BD09MC, BD09)
	if err != nil {
		t.Fatalf("Convert BD09MC to BD09 error = %v", err)
	}
	if Distance(bd, roundTrip) > 1 {
		t.Fatalf("BD09MC round trip distance = %.2fm, coord = %#v", Distance(bd, roundTrip), roundTrip)
	}
}

func TestBD09MCConvenienceFunctionsMatchConvert(t *testing.T) {
	wgs := Coord{Lng: 116.397389, Lat: 39.908722}
	gcj := WGS84ToGCJ02(wgs)
	bd := WGS84ToBD09(wgs)

	mcFromWGS, err := Convert(wgs, WGS84, BD09MC)
	if err != nil {
		t.Fatalf("Convert WGS84 to BD09MC error = %v", err)
	}
	if got := WGS84ToBD09MC(wgs); got != mcFromWGS {
		t.Fatalf("WGS84ToBD09MC = %#v, want %#v", got, mcFromWGS)
	}
	if got := GCJ02ToBD09MC(gcj); Distance(BD09MCToBD09(got), bd) > 1 {
		t.Fatalf("GCJ02ToBD09MC round trip through BD09 distance = %.2fm", Distance(BD09MCToBD09(got), bd))
	}
	if got := BD09ToBD09MC(bd); got != mcFromWGS {
		t.Fatalf("BD09ToBD09MC = %#v, want %#v", got, mcFromWGS)
	}
	if got := BD09MCToBD09(mcFromWGS); Distance(got, bd) > 1 {
		t.Fatalf("BD09MCToBD09 distance = %.2fm", Distance(got, bd))
	}
	if got := BD09MCToGCJ02(mcFromWGS); Distance(got, gcj) > 1 {
		t.Fatalf("BD09MCToGCJ02 distance = %.2fm", Distance(got, gcj))
	}
	if got := BD09MCToWGS84(mcFromWGS); Distance(got, wgs) > 2 {
		t.Fatalf("BD09MCToWGS84 distance = %.2fm", Distance(got, wgs))
	}
}

func TestConvertBD09MCNegativeAndClamped(t *testing.T) {
	bd := Coord{Lng: -200, Lat: -90}
	mc, err := Convert(bd, BD09, BD09MC)
	if err != nil {
		t.Fatalf("Convert clamped BD09 to BD09MC error = %v", err)
	}
	roundTrip, err := Convert(mc, BD09MC, BD09)
	if err != nil {
		t.Fatalf("Convert clamped BD09MC to BD09 error = %v", err)
	}
	if roundTrip.Lng < 159.9 || roundTrip.Lng > 160.1 {
		t.Fatalf("roundTrip longitude = %.6f, want looped longitude near 160", roundTrip.Lng)
	}
	if roundTrip.Lat < -74.1 || roundTrip.Lat > -73.9 {
		t.Fatalf("roundTrip latitude = %.6f, want clamped latitude near -74", roundTrip.Lat)
	}
}

func TestConvertBD09MCChainedSystems(t *testing.T) {
	wgs := Coord{Lng: 116.397389, Lat: 39.908722}
	mc, err := Convert(wgs, WGS84, BD09MC)
	if err != nil {
		t.Fatalf("Convert WGS84 to BD09MC error = %v", err)
	}
	back, err := Convert(mc, BD09MC, WGS84)
	if err != nil {
		t.Fatalf("Convert BD09MC to WGS84 error = %v", err)
	}
	if Distance(wgs, back) > 2 {
		t.Fatalf("WGS84 <-> BD09MC round trip distance = %.2fm, coord = %#v", Distance(wgs, back), back)
	}
}

func TestConvertSameSupportedType(t *testing.T) {
	coord := Coord{Lng: 116.397389, Lat: 39.908722}
	got, err := Convert(coord, BD09MC, BD09MC)
	if err != nil {
		t.Fatalf("Convert same supported type error = %v", err)
	}
	if got != coord {
		t.Fatalf("Convert same supported type = %#v, want %#v", got, coord)
	}
}

func TestConvertRejectsUnsupportedTypeEvenWhenEqual(t *testing.T) {
	coord := Coord{Lng: 1, Lat: 2}
	_, err := Convert(coord, CoordType("UNKNOWN"), CoordType("UNKNOWN"))
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("Convert unsupported same type error = %v", err)
	}
}

func TestConvertRejectsUnsupportedPair(t *testing.T) {
	_, err := Convert(Coord{}, WGS84, CoordType("UNKNOWN"))
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("Convert unsupported pair error = %v", err)
	}
}

func TestInChinaAndDistance(t *testing.T) {
	outside := Coord{Lng: -73.9857, Lat: 40.7484}
	if got := WGS84ToGCJ02(outside); got != outside {
		t.Fatalf("WGS84ToGCJ02 outside China = %#v", got)
	}
	if !InChina(116.3, 39.9) || InChina(-73, 40) {
		t.Fatal("InChina boundary check failed")
	}
	if d := Distance(Coord{Lng: 0, Lat: 0}, Coord{Lng: 0, Lat: 1}); math.Abs(d-111195) > 200 {
		t.Fatalf("Distance = %.2f", d)
	}
}

func TestInChinaBoundaryValues(t *testing.T) {
	tests := []struct {
		name string
		lng  float64
		lat  float64
		want bool
	}{
		{name: "minimum included", lng: 72.004, lat: 0.8293, want: true},
		{name: "maximum included", lng: 137.8347, lat: 55.8271, want: true},
		{name: "longitude below", lng: 72.0039, lat: 30, want: false},
		{name: "latitude above", lng: 100, lat: 55.8272, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InChina(tt.lng, tt.lat); got != tt.want {
				t.Fatalf("InChina(%f, %f) = %v, want %v", tt.lng, tt.lat, got, tt.want)
			}
		})
	}
}

func TestDistanceAntipodalIsFinite(t *testing.T) {
	d := Distance(Coord{Lng: 0, Lat: 0}, Coord{Lng: 180, Lat: 0})
	if math.IsNaN(d) || math.IsInf(d, 0) {
		t.Fatalf("Distance returned non-finite value: %v", d)
	}
	want := math.Pi * earthRadiusM
	if math.Abs(d-want) > 1 {
		t.Fatalf("Distance antipodal = %.2f, want %.2f", d, want)
	}
}

func distanceApprox(a, b Coord) float64 {
	return Distance(a, b)
}
