// Package boolean provides boolean helpers.
package boolean

// Negate returns the logical negation of b.
func Negate(b bool) bool { return !b }

// ToInt returns 1 for true and 0 for false.
func ToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// And returns true only when all inputs are true.
func And(bs ...bool) bool {
	for _, b := range bs {
		if !b {
			return false
		}
	}
	return true
}

// Or returns true when any input is true.
func Or(bs ...bool) bool {
	for _, b := range bs {
		if b {
			return true
		}
	}
	return false
}
