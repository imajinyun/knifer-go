//go:generate go run ./bin/api_snapshot.go
//go:generate go run ./bin/toolsgen -out docs/api/tools.json -markdown docs/api/tools.md

// Package knifer is the root package of the go-knifer utility toolkit.
//
// This module is split into 48 public subpackages by domain. Import only the
// packages you need. The subpackages are grouped below for navigation:
//
// String & text:
//
//	vstr    strings and text similarity helpers
//	vregex  regular expressions
//	vtpl    html/template rendering (TemPLate)
//	vurl    URL/URI parsing, escaping, query building
//
// Collections & data structures:
//
//	vslice  slices
//	vmap    maps
//	vset    sets
//	vobj    object-level helpers
//	vblf    bloom filters (BLoom Filter)
//	vcache  generic caches (FIFO/LRU/LFU/Timed)
//	vbean   struct/map mapping and copying
//
// Primitives & conversion:
//
//	vbool   booleans
//	vnum    numeric helpers
//	vconv   permissive type conversion
//	vdate   date/time
//	vref    reflection
//
// Encoding & serialization:
//
//	vcodec  Base64/Hex
//	vcsv    CSV reading/writing
//	vimg    raster images and graphical captchas
//	vjson   JSON
//	vxml    XML
//	vhash   non-cryptographic hashes
//
// Networking & communication:
//
//	vhttp   standard-library HTTP client/server
//	vresty  Resty-based HTTP client
//	vmail   email message construction, MIME attachments, and SMTP sending
//	vskt    sockets (SocKeT)
//	vnet    IP/port/interface utilities
//
// Security & matching:
//
//	vcrypto cryptography and digests
//	vjwt    JWT sign/verify
//	vmask   data masking (desensitization)
//	vpass   password strength analysis
//	vdfa    DFA word-tree text matching
//
// Tasks & concurrency:
//
//	vjob    job orchestration
//	vcron   cron scheduling
//	vsem    semaphores (SEMaphore)
//	vrand   randomness
//
// IO & files:
//
//	vfile   file and IO helpers
//	vzip    archive/compression
//	vpoi    office documents (Excel)
//
// Runtime & system:
//
//	vsys    system information (SYStem)
//	vlog    logging
//	verr    error handling, panic recovery, stacks (errx)
//	vconf   configuration
//
// Identity & misc:
//
//	vid     generated IDs (UUID/Snowflake/ObjectId/NanoId)
//	vident  legal identity numbers (ID cards)
//	vform   form and input validators
//	vver    version comparison (VERsion)
//	vdb     database/sql helpers
//
// Example:
//
//	import "github.com/imajinyun/go-knifer/vstr"
//	import "github.com/imajinyun/go-knifer/vhttp"
//
// Subpackages are independent from each other. The root package exposes no
// business APIs; it only defines the cross-cutting error contract (ErrCode,
// Error, CodeCarrier, CodeOf and the New/Wrap/Errorf constructors) that
// subpackages may use to classify failures consistently.
//
// Callers can match failures by code regardless of the originating subpackage:
//
//	if errors.Is(err, knifer.ErrCodeInvalidInput) { ... }
//	if code, ok := knifer.CodeOf(err); ok { ... }
//
// Subpackages that participate should return *knifer.Error or implement
// CodeCarrier on their existing error types/sentinels while wrapping any
// underlying cause so the standard error chain is preserved.
//
// The project follows an internal implementation plus public facade layout:
// concrete implementations live in internal/* packages, while application code
// should import the public v* packages. This keeps domain boundaries explicit
// and allows internal implementations to evolve without exposing every helper as
// public API.
package knifer
