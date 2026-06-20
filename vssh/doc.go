// Package vssh provides provider-neutral SSH command and SFTP transfer helpers.
//
// The package requires provider injection. It does not open network connections,
// execute shell commands, read credentials, parse keys, or touch local filesystem
// paths by default.
package vssh
