// Package tokenize implements provider-neutral text tokenization primitives.
//
// The package defines request and response contracts for tokenization and
// keyword extraction providers. It does not import dictionaries, segment text,
// rank keywords, open network connections, read credentials, or touch local
// filesystem paths by default.
package tokenize
