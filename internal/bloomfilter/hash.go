package bloomfilter

// The non-cryptographic string hash algorithms used by Bloom filters live in
// internal/hash as the single source of truth. These aliases keep the existing
// call sites within this package unchanged.

import hashimpl "github.com/imajinyun/knifer-go/internal/hash"

var (
	// RsHash implements the RS algorithm.
	RsHash = hashimpl.RsHash
	// JsHash implements the JS algorithm.
	JsHash = hashimpl.JsHash
	// PjwHash implements the PJW algorithm.
	PjwHash = hashimpl.PjwHash
	// ElfHash implements the ELF algorithm.
	ElfHash = hashimpl.ElfHash
	// BkdrHash implements the BKDR algorithm.
	BkdrHash = hashimpl.BkdrHash
	// SdbmHash implements the SDBM algorithm.
	SdbmHash = hashimpl.SdbmHash
	// DjbHash implements the DJB algorithm.
	DjbHash = hashimpl.DjbHash
	// ApHash implements the AP algorithm.
	ApHash = hashimpl.ApHash
	// FnvHashString implements the improved 32-bit FNV-1 algorithm for strings.
	FnvHashString = hashimpl.FnvHashString
	// HfHash implements the HF hash algorithm.
	HfHash = hashimpl.HfHash
	// HfIpHash implements the HFIP hash algorithm.
	HfIpHash = hashimpl.HfIpHash
	// TianlHash implements the TianL hash algorithm.
	TianlHash = hashimpl.TianlHash
	// JavaDefaultHash simulates Java String.hashCode.
	JavaDefaultHash = hashimpl.JavaDefaultHash
)
