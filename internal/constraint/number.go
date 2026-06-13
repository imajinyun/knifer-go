package constraint

// Signed permits all built-in signed integer types.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned permits all built-in unsigned integer types.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Integer permits all built-in signed and unsigned integer types.
type Integer interface {
	Signed | Unsigned
}

// Float permits all built-in floating-point types.
type Float interface {
	~float32 | ~float64
}

// Complex permits all built-in complex numeric types.
type Complex interface {
	~complex64 | ~complex128
}

// Number permits all built-in integer and floating-point types.
type Number interface {
	Integer | Float
}
