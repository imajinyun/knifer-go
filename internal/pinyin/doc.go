// Package pinyin implements provider-neutral Chinese-to-pinyin primitives.
//
// The package defines request and response contracts for pinyin conversion and
// initials extraction providers. It does not import dictionaries, tokenize text,
// open network connections, read credentials, or touch local filesystem paths by
// default.
package pinyin
