package vgeo_test

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vgeo"
)

func ExampleWGS84ToGCJ02() {
	gcj := vgeo.WGS84ToGCJ02(vgeo.Coord{Lng: 116.397389, Lat: 39.908722})
	fmt.Printf("%.6f %.6f\n", gcj.Lng, gcj.Lat)
	// Output: 116.403633 39.910125
}

func ExampleDistance() {
	d := vgeo.Distance(
		vgeo.Coord{Lng: 116.397389, Lat: 39.908722},
		vgeo.Coord{Lng: 121.4737, Lat: 31.2304},
	)
	fmt.Printf("%.0f\n", d/1000)
	// Output: 1068
}

func ExampleConvert() {
	coord, err := vgeo.Convert(
		vgeo.Coord{Lng: 116.397389, Lat: 39.908722},
		vgeo.WGS84,
		vgeo.BD09,
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%.6f %.6f\n", coord.Lng, coord.Lat)
	// Output: 116.410005 39.916465
}

func ExampleGCJ02ToBD09() {
	bd := vgeo.GCJ02ToBD09(vgeo.Coord{Lng: 116.403633, Lat: 39.910125})
	fmt.Printf("%.6f %.6f\n", bd.Lng, bd.Lat)
	// Output: 116.410006 39.916465
}

func ExampleBD09ToGCJ02() {
	gcj := vgeo.BD09ToGCJ02(vgeo.Coord{Lng: 116.410008, Lat: 39.916471})
	fmt.Printf("%.6f %.6f\n", gcj.Lng, gcj.Lat)
	// Output: 116.403636 39.910131
}

func ExampleBD09ToWGS84() {
	wgs := vgeo.BD09ToWGS84(vgeo.Coord{Lng: 116.410008, Lat: 39.916471})
	fmt.Printf("%.6f %.6f\n", wgs.Lng, wgs.Lat)
	// Output: 116.397392 39.908727
}

func ExampleInChina() {
	fmt.Println(vgeo.InChina(116.397389, 39.908722))
	fmt.Println(vgeo.InChina(-73.9857, 40.7484))
	// Output:
	// true
	// false
}
