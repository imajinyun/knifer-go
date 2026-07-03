package vgeo_test

import (
	"errors"
	"testing"

	knifer "github.com/imajinyun/knifer-go"
	"github.com/imajinyun/knifer-go/vgeo"
)

func TestConvertBD09MCRoundTrip(t *testing.T) {
	bd := vgeo.Coord{Lng: 116.410005, Lat: 39.916465}
	mc, err := vgeo.Convert(bd, vgeo.BD09, vgeo.BD09MC)
	if err != nil {
		t.Fatalf("Convert BD09 to BD09MC error = %v", err)
	}
	roundTrip, err := vgeo.Convert(mc, vgeo.BD09MC, vgeo.BD09)
	if err != nil {
		t.Fatalf("Convert BD09MC to BD09 error = %v", err)
	}
	if d := vgeo.Distance(bd, roundTrip); d > 1 {
		t.Fatalf("BD09MC round trip distance = %.2fm, coord = %#v", d, roundTrip)
	}
}

func TestBD09MCConvenienceFunctions(t *testing.T) {
	wgs := vgeo.Coord{Lng: 116.397389, Lat: 39.908722}
	mc := vgeo.WGS84ToBD09MC(wgs)
	if fromConvert, err := vgeo.Convert(wgs, vgeo.WGS84, vgeo.BD09MC); err != nil {
		t.Fatalf("Convert WGS84 to BD09MC error = %v", err)
	} else if fromConvert != mc {
		t.Fatalf("WGS84ToBD09MC = %#v, want %#v", mc, fromConvert)
	}
	if d := vgeo.Distance(vgeo.BD09MCToWGS84(mc), wgs); d > 2 {
		t.Fatalf("BD09MCToWGS84 distance = %.2fm", d)
	}
	if d := vgeo.Distance(vgeo.BD09MCToGCJ02(mc), vgeo.WGS84ToGCJ02(wgs)); d > 1 {
		t.Fatalf("BD09MCToGCJ02 distance = %.2fm", d)
	}
	if d := vgeo.Distance(vgeo.BD09MCToBD09(vgeo.BD09ToBD09MC(vgeo.WGS84ToBD09(wgs))), vgeo.WGS84ToBD09(wgs)); d > 1 {
		t.Fatalf("BD09MC BD09 round trip distance = %.2fm", d)
	}
	if d := vgeo.Distance(vgeo.BD09MCToBD09(vgeo.GCJ02ToBD09MC(vgeo.WGS84ToGCJ02(wgs))), vgeo.WGS84ToBD09(wgs)); d > 1 {
		t.Fatalf("GCJ02ToBD09MC round trip distance = %.2fm", d)
	}
}

func TestConvertRejectsUnknownCoordType(t *testing.T) {
	_, err := vgeo.Convert(vgeo.Coord{}, vgeo.CoordType("UNKNOWN"), vgeo.CoordType("UNKNOWN"))
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("Convert unknown same type error = %v", err)
	}
}
