# vident Quickstart

`vident` provides validation, conversion, birthdate/age/gender/region parsing, and masking helpers for mainland China resident ID cards, Hong Kong/Macau/Taiwan documents, and unified social credit codes.

## Which helper should I use?

Choose helpers by document type and by whether you need validation, derived fields, or redaction.

| Need | Use | Notes |
| --- | --- | --- |
| Validate mainland China IDs | `IsValidIDCard`, `IsValidIDCard18`, `IsValidIDCard15` | Use before parsing derived fields or accepting a document number. |
| Convert between 15- and 18-digit IDs | `Convert15To18`, `Convert18To15` | Check the returned `ok` flag before using the converted value. |
| Parse birth date and age | `BirthString`, `BirthDate`, `Age`, `AgeAt`, `AgeWithOptions` | Prefer `AgeAt` or clock options in tests for deterministic results. |
| Parse gender and region fields | `GenderOf`, `Province`, `CityCode`, `DistrictCode`, `ParseIDCard` | Region names depend on the package's region data and may not represent current administrative boundaries. |
| Validate HK/Macau/Taiwan documents | `ParseRegionCard`, `IsValidTWIDCard`, `IsValidHKIDCard` | Use document-specific helpers when the expected region is known. |
| Validate unified social credit codes | `IsValidCreditCode`, `ParseCreditCode` | Checks GB 32100-style character, length, and check-digit rules, then splits the registration fields. |
| Mask document numbers | `Hide` | Redact before logging, displaying, or exporting identifiers. |
| Inject matchers/parsers for tests | `WithDigitsMatcher`, region matcher options, birth and age options | Keeps edge cases deterministic without depending on process time or external data. |

## Identity-data safety checklist

- Treat all document numbers as sensitive personal data. Mask before logs, screenshots, metrics, and support exports.
- Always check the boolean return before using a converted, parsed, or derived value.
- Treat unified social credit codes as organization identifiers that may still be sensitive in supplier, customer, and account workflows.
- Use deterministic clocks (`AgeAt` or `WithAgeClock`) in tests and batch jobs that require reproducible output.
- Validate the expected document region instead of accepting every supported regional format by default.
- Do not treat parsed region or age fields as proof of identity; they are derived from a claimed document number.
- Review administrative-region data freshness when region names are user-visible or regulatory decisions depend on them.

## When not to use vident

- Use an official identity verification provider when legal identity, fraud prevention, or KYC requirements apply.
- Use a privacy-preserving token or internal subject ID when the raw document number is not required.
- Use domain-specific validation when product policy accepts only one region or one document class.
- Avoid storing raw IDs if masked, hashed, encrypted, or tokenized forms satisfy the workflow.

## Related packages

- Use `vform` when identity documents are one field in a larger validation workflow.
- Use `vstr` when input normalization, trimming, or masking-adjacent string cleanup is needed first.
- Use `vmask` when validated identity data must be redacted for logs, UI, or diagnostics.

## Benchmarks and trade-offs

Validation and parsing are mostly string operations, but batch imports should still measure matcher and birth-date parsing costs:

```bash
go test -bench=. -benchmem -run=^$ ./internal/identity ./vident
```

Convenience helpers reduce repeated checksum, birth-date, and region parsing code. The trade-off is policy ambiguity: callers still need to decide which document classes are allowed, how stale region data is handled, and where PII may flow.

## FAQ

### Does a valid checksum prove the person is real?

No. Validation only checks format, date, region, and checksum rules. It does not verify ownership, current validity, or legal identity.

### Why should age calculations inject time?

Age changes with the current date. `AgeAt` and `WithAgeClock` keep tests, reports, and audits reproducible.

### Can I log masked IDs?

Masked IDs are safer than raw values, but they may still be personal data depending on policy and context. Keep logs minimal and avoid raw document numbers.

## Validate and convert ID card numbers

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vident"
)

func main() {
	id18, ok := vident.Convert15To18("130503670401001")
	fmt.Println(id18, ok)

	id15, ok := vident.Convert18To15("11010519491231002X")
	fmt.Println(id15, ok)
	fmt.Println(vident.IsValidIDCard("11010519491231002X"))
	fmt.Println(vident.IsValidIDCard18WithIgnoreCase("11010519491231002x", true))
}
```

## Parse birthdates and ages

```go
package main

import (
	"fmt"
	"time"

	"github.com/imajinyun/knifer-go/vident"
)

func main() {
	id := "11010519491231002X"
	birth, ok := vident.BirthDate(id)
	fmt.Println(birth.Format("2006-01-02"), ok)

	age, ok := vident.AgeAt(id, time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local))
	fmt.Println(age, ok)
	fmt.Println(vident.IsValidBirthday("19491231"))

	age, ok = vident.AgeWithOptions(id, vident.WithAgeClock(func() time.Time {
		return time.Date(2024, 12, 30, 0, 0, 0, 0, time.Local)
	}))
	fmt.Println(age, ok)
}
```

Use `BirthDateWithOptions`, `BirthStringWithOptions`, `IsValidBirthdayWithOptions`, `WithBirthLocation`, `WithBirthParser`, and `WithBirthDigitsMatcher` when tests or import jobs need a fixed parser, a fixed location, or a custom digit policy.

## Parse gender and region information

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vident"
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
	fmt.Println(vident.ProvinceCode(id))
	fmt.Println(vident.CityCode(id))
	fmt.Println(vident.DistrictCode(id))
	year, _ := vident.Year(id)
	month, _ := vident.Month(id)
	day, _ := vident.Day(id)
	fmt.Println(year, month, day)
}
```

## Validate Hong Kong/Macau/Taiwan documents and mask values

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vident"
)

func main() {
	region, ok := vident.ParseRegionCard("A123456(3)")
	fmt.Println(region.Region, region.Valid, ok)

	macau, ok := vident.ParseRegionCardWithOptions("1571234(5)")
	fmt.Println(macau.Region, macau.Valid, ok)

	fmt.Println(vident.IsValidHKIDCard("A123456(3)"))
	fmt.Println(vident.IsValidTWIDCard("A123456789"))
	fmt.Println(vident.IsValidTWIDCardWithOptions(
		"A123456789",
		vident.WithTWCardMatcher(func(card string) bool { return len(card) == 10 }),
	))
	fmt.Println(vident.Hide("11010519491231002X", 6, 14))

	rejected := vident.IsValidIDCard18WithOptions(
		"11010519491231002X",
		vident.WithDigitsMatcher(func(string) bool { return false }),
	)
	fmt.Println(rejected)
}
```

Use `WithHKCardMatcher`, `WithMacauCardMatcher`, and `WithTWCardMatcher` when a product has stricter document-format policy than the package defaults, or when tests need to force region-specific reject paths.

## Validate unified social credit codes

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vident"
)

func main() {
	code := "91350211M000100Y46"
	fmt.Println(vident.IsValidCreditCode(code))

	info, err := vident.ParseCreditCode(code)
	if err != nil {
		panic(err)
	}

	fmt.Println(info.AdminDept, info.OrgCategory, info.RegionCode)
	fmt.Println(info.OrgCode, info.CheckDigit)
}
```

`ParseCreditCode` normalizes surrounding space and letter case before validation. It validates format and checksum only; it does not confirm that the organization exists, is active, or owns the submitted code.
