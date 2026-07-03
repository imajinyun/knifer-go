# vgeo Quickstart

`vgeo` provides coordinate conversion helpers for WGS-84, GCJ-02, and BD-09 longitude/latitude coordinates, plus Haversine distance.

## Which helper should I use?

Choose explicit conversions when the source and target systems are known. Use `Convert` when the coordinate system is a runtime value.

| Need | Use | Notes |
| --- | --- | --- |
| GPS to China map coordinates | `WGS84ToGCJ02` | Applies the GCJ-02 offset only inside a rough China bounding box. |
| China map coordinates back to GPS | `GCJ02ToWGS84` | Uses the common approximate inverse; it is suitable for meter-level utility workflows, not surveying. |
| Baidu map coordinates | `GCJ02ToBD09`, `BD09ToGCJ02`, `WGS84ToBD09`, `BD09ToWGS84` | BD-09 conversions compose through GCJ-02 where needed. |
| Runtime coordinate-system routing | `Convert` | Supports WGS-84, GCJ-02, BD-09, and BD-09 Mercator; returns `ErrCodeInvalidInput` for unknown coordinate systems. |
| Distance between coordinates | `Distance` | Returns meters using the Haversine formula. |
| Offset decision checks | `InChina` | Uses a coarse bounding box, not administrative-boundary geometry. |

## Coordinate safety checklist

- Store the coordinate system with every coordinate; longitude and latitude alone are ambiguous.
- Do not use these helpers for surveying, legal boundary, or high-precision geodesy workflows.
- Treat GCJ-02 reverse conversion as an approximation.
- Keep longitude/latitude order visible in code: `Coord{Lng: ..., Lat: ...}`.
- Add tolerance-based tests for conversion output because floating-point rounding can vary at the last decimal.

## When not to use vgeo

- Use a GIS or geodesy library for projections, datum transforms, route planning, or administrative-boundary checks.
- BD-09 Mercator support is intended for Baidu map interoperability, not general-purpose projection work.
- Do not assume `InChina` is an exact policy boundary.

## Related packages

- Use `vcodec` when coordinates need to be encoded for compact transport.
- Use `vjson` when serializing coordinate payloads with explicit system metadata.
- Use `vnum` for numeric formatting or rounding at presentation boundaries.

## Benchmarks and trade-offs

Coordinate conversion is CPU-only and allocation-light. Benchmark hot paths with representative batches:

```bash
go test -bench=. -benchmem -run=^$ ./internal/geo ./vgeo
```

The helpers favor small, dependency-free utility behavior over projection breadth. That keeps `vgeo` usable in core packages, but callers needing GIS-grade capabilities should use a dedicated library.

## FAQ

### Is GCJ-02 to WGS-84 exact?

No. The reverse helper uses the common approximate inverse. It is practical for display and integration utilities, but not a precision geodesy contract.

### Why does `WGS84ToGCJ02` return unchanged coordinates outside China?

GCJ-02 offsets are only applied for coordinates inside the rough China bounding box used by `InChina`.

### Does `Distance` account for terrain or road routes?

No. It returns a spherical great-circle distance in meters.

## Convert GPS to GCJ-02

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vgeo"
)

func main() {
	gcj := vgeo.WGS84ToGCJ02(vgeo.Coord{Lng: 116.397389, Lat: 39.908722})
	fmt.Printf("%.6f %.6f\n", gcj.Lng, gcj.Lat)
}
```

## Convert through runtime coordinate types

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vgeo"
)

func main() {
	coord, err := vgeo.Convert(
		vgeo.Coord{Lng: 116.397389, Lat: 39.908722},
		vgeo.WGS84,
		vgeo.BD09,
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%.6f %.6f\n", coord.Lng, coord.Lat)
}
```

## Convert Baidu BD-09 coordinates

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vgeo"
)

func main() {
	wgs := vgeo.Coord{Lng: 116.397389, Lat: 39.908722}
	bd := vgeo.WGS84ToBD09(wgs)
	back := vgeo.BD09ToWGS84(bd)

	fmt.Printf("%.6f %.6f\n", bd.Lng, bd.Lat)
	fmt.Printf("%.6f %.6f\n", back.Lng, back.Lat)
}
```

BD-09 helpers are intended for map-provider interoperability. Round trips may differ by a few meters because GCJ-02 reverse conversion is approximate.

## Measure coordinate distance

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vgeo"
)

func main() {
	beijing := vgeo.Coord{Lng: 116.397389, Lat: 39.908722}
	shanghai := vgeo.Coord{Lng: 121.4737, Lat: 31.2304}
	fmt.Printf("%.0f km\n", vgeo.Distance(beijing, shanghai)/1000)
}
```
