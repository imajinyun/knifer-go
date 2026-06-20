// Package ssh implements provider-neutral SSH and SFTP transfer primitives.
//
// The package defines request and response contracts for command execution,
// remote listing, download, and upload providers. It does not open network
// connections, execute shell commands, read credentials, parse keys, or touch
// local filesystem paths by default.
package ssh
