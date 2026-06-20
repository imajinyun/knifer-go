# vident Quickstart

`vident` provides validation, conversion, birthdate/age/gender/region parsing, and masking helpers for mainland China resident ID cards and Hong Kong/Macau/Taiwan documents.

## Validate and convert ID card numbers

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vident"
)

func main() {
	id18, ok := vident.Convert15To18("130503670401001")
	fmt.Println(id18, ok)

	id15, ok := vident.Convert18To15("11010519491231002X")
	fmt.Println(id15, ok)
	fmt.Println(vident.IsValidIDCard("11010519491231002X"))
}
```

## Parse birthdates and ages

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/go-knifer/vident"
)

func main() {
	id := "11010519491231002X"
	birth, ok := vident.BirthDate(id)
	fmt.Println(birth.Format("2006-01-02"), ok)

	age, ok := vident.AgeAt(id, time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local))
	fmt.Println(age, ok)
	fmt.Println(vident.IsValidBirthday("19491231"))
}
```

## Parse gender and region information

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vident"
)

func main() {
	id := "11010519491231002X"
	info, ok := vident.ParseIDCard(id)
	if !ok {
		panic("invalid id card")
	}

	fmt.Println(info.Province, info.CityCode, info.DistrictCode)
	fmt.Println(info.Gender == vident.GenderFemale)
	fmt.Println(vident.Province(id))
}
```

## Validate Hong Kong/Macau/Taiwan documents and mask values

```go
package main

import (
	"fmt"

	"github.com/imajinyun/go-knifer/vident"
)

func main() {
	region, ok := vident.ParseRegionCard("A123456(3)")
	fmt.Println(region.Region, region.Valid, ok)

	fmt.Println(vident.IsValidTWIDCard("A123456789"))
	fmt.Println(vident.Hide("11010519491231002X", 6, 14))
}
```
