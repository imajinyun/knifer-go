// Package ftp implements provider-neutral FTP transfer primitives.
//
// The package defines request and response contracts for list, download, and
// upload providers. It does not open network connections, read credentials, or
// touch local filesystem paths by default.
package ftp
