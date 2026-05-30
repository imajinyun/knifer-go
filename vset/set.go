package vset

import setimpl "github.com/imajinyun/go-knifer/internal/sets"

// Int is a set of int values.
type Int = setimpl.Int

// Int32 is a set of int32 values.
type Int32 = setimpl.Int32

// Int64 is a set of int64 values.
type Int64 = setimpl.Int64

// Uint is a set of uint values.
type Uint = setimpl.Uint

// Uint32 is a set of uint32 values.
type Uint32 = setimpl.Uint32

// Uint64 is a set of uint64 values.
type Uint64 = setimpl.Uint64

// String is a set of string values.
type String = setimpl.String

// NewInt creates an int set.
func NewInt(items ...int) Int { return setimpl.NewInt(items...) }

// NewInt32 creates an int32 set.
func NewInt32(items ...int32) Int32 { return setimpl.NewInt32(items...) }

// NewInt64 creates an int64 set.
func NewInt64(items ...int64) Int64 { return setimpl.NewInt64(items...) }

// NewUint creates a uint set.
func NewUint(items ...uint) Uint { return setimpl.NewUint(items...) }

// NewUint32 creates a uint32 set.
func NewUint32(items ...uint32) Uint32 { return setimpl.NewUint32(items...) }

// NewUint64 creates a uint64 set.
func NewUint64(items ...uint64) Uint64 { return setimpl.NewUint64(items...) }

// NewString creates a string set.
func NewString(items ...string) String { return setimpl.NewString(items...) }
