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

func TestConvertUnsupported(t *testing.T) {
	_, err := Convert(Coord{}, BD09MC, WGS84)
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("Convert unsupported error = %v", err)
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

func distanceApprox(a, b Coord) float64 {
	return Distance(a, b)
}
