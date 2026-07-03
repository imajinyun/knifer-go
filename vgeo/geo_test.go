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

func TestConvertRejectsUnknownCoordType(t *testing.T) {
	_, err := vgeo.Convert(vgeo.Coord{}, vgeo.CoordType("UNKNOWN"), vgeo.CoordType("UNKNOWN"))
	if !errors.Is(err, knifer.ErrCodeInvalidInput) {
		t.Fatalf("Convert unknown same type error = %v", err)
	}
}
