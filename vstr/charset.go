package vstr

import strimpl "github.com/imajinyun/knifer-go/internal/str"

// ToUTF8 converts data from the named charset to UTF-8.
func ToUTF8(data []byte, from string) ([]byte, error) { return strimpl.ToUTF8(data, from) }

// FromUTF8 converts UTF-8 data to the named charset.
func FromUTF8(data []byte, to string) ([]byte, error) { return strimpl.FromUTF8(data, to) }
