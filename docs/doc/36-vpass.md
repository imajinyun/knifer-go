# vpass Quickstart

`vpass` provides password strength analysis, scoring, strength levels, and shortcut checks for strong or weak passwords.

## Golden path APIs

The first-choice API set for this facade is kept in sync with `ai-context.json` and the generated tools catalog.

- `Analyze`
- `IsStrong`
- `IsWeak`
- `Score`
- `StrengthOf`

## Which helper should I use?

| Goal | Start with | Notes |
| --- | --- | --- |
| Get the full rule-level result | `Analyze` | Returns `Analysis` with score and strength so callers can show feedback. |
| Need only the numeric score | `Score` | Use when policy stores or compares score thresholds directly. |
| Need only a strength bucket | `StrengthOf` | Returns `StrengthVeryWeak` through `StrengthVeryStrong`. |
| Fast accept/reject predicates | `IsStrong`, `IsWeak` | Useful for simple UI or policy gates, but still pair with server-side policy. |
| Render strength labels | `Strength.String()` | Handles known values and `StrengthUnknown` consistently. |

## Password policy checklist

- Treat strength analysis as one signal, not a complete password policy.
- Check leaked-password lists, minimum length, account-specific context, and rate limits outside `vpass`.
- Run checks server-side even if the UI also calls equivalent logic.
- Do not log raw passwords, scores tied to a user, or password analysis details in production logs.
- Prefer passphrases and password managers; avoid policies that force predictable substitutions only.
- Use `Analyze` when users need actionable feedback instead of only rejecting with `IsWeak`.

## Analyze password strength

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vpass"
)

func main() {
	analysis := vpass.Analyze("G0-Knifer#Pass2026")
	fmt.Println(analysis.Score)
	fmt.Println(analysis.Strength == vpass.StrengthVeryStrong)
}
```

## Get scores and strength levels

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vpass"
)

func main() {
	fmt.Println(vpass.Score("password"))
	fmt.Println(vpass.StrengthOf("password") == vpass.StrengthVeryWeak)
	fmt.Println(vpass.StrengthUnknown.String())
}
```

## Quickly detect strong and weak passwords

```go
package main

import (
	"fmt"

	"github.com/imajinyun/knifer-go/vpass"
)

func main() {
	fmt.Println(vpass.IsStrong("G0-Knifer#Pass2026"))
	fmt.Println(vpass.IsWeak("12345"))
}
```

## When not to use vpass

- Use a full authentication or identity system for password hashing, storage, reset flows, MFA, and breach checks.
- Use a leaked-password corpus check when the security requirement is to reject compromised secrets.
- Avoid treating `IsStrong` as proof that a password is safe for every account or threat model.
- Do not use strength scoring for API keys or generated tokens; use entropy and randomness requirements instead.

## Related packages

- Use `vcrypto` when password workflows need dedicated hashing, key derivation, or cryptographic verification.
- Use `vrand` when generating secure temporary passwords, salts, or reset token bytes.
- Use `vform` when password strength checks are part of a larger signup or credential validation form.

## Benchmarks and trade-offs

- `Score`, `StrengthOf`, `IsStrong`, and `IsWeak` are convenience wrappers around analysis-style logic, so call `Analyze` once if a flow needs multiple fields.
- Strength scoring is fast enough for form validation, but breach-list checks and account-context checks are intentionally out of scope.
- Bucketed strength levels are easy to explain to users, while numeric scores give applications finer policy control.
- More feedback can improve password choice but can also reveal policy details; tune UI messaging for your risk model.

## FAQ

### Is `IsStrong` enough before accepting a password?

No. Pair it with minimum length, breached-password checks, rate limiting, secure hashing, and account-specific policy.

### Should I call `Analyze` or `Score`?

Call `Analyze` when you need both score and strength or want to show feedback. Call `Score` when a numeric threshold is the only output needed.

### Can `vpass` store or hash passwords?

No. Use a password hashing library such as bcrypt, scrypt, or Argon2 through your authentication layer.

### Do vpass helpers return errors?

No. `vpass` analyzes a provided string and returns deterministic scores or strength buckets. Handle input errors, password policy failures, and storage errors in the authentication layer.

### Does a very strong password remove the need for MFA?

No. Password strength and MFA defend against different failure modes and should be considered separately.
