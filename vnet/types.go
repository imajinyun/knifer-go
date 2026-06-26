package vnet

import netimpl "github.com/imajinyun/knifer-go/internal/net"

const (
	LocalIP         = netimpl.LocalIP
	IPSplitMark     = netimpl.IPSplitMark
	IPMaskSplitMark = netimpl.IPMaskSplitMark
	IPMaskMax       = netimpl.IPMaskMax
	PortRangeMin    = netimpl.PortRangeMin
	PortRangeMax    = netimpl.PortRangeMax
	SSL             = netimpl.SSL
	SSLv2           = netimpl.SSLv2
	SSLv3           = netimpl.SSLv3
	TLS             = netimpl.TLS
	TLSv1           = netimpl.TLSv1
	TLSv11          = netimpl.TLSv11
	TLSv12          = netimpl.TLSv12
	TLSv13          = netimpl.TLSv13
)

type (
	TLSConfigBuilder     = netimpl.TLSConfigBuilder
	Dialer               = netimpl.Dialer
	ConnectOption        = netimpl.ConnectOption
	PingOption           = netimpl.PingOption
	ResolveOption        = netimpl.ResolveOption
	AddressOption        = netimpl.AddressOption
	PortOption           = netimpl.PortOption
	InterfaceOption      = netimpl.InterfaceOption
	TLSFileOption        = netimpl.TLSFileOption
	UploadSetting        = netimpl.UploadSetting
	UploadSaveOption     = netimpl.UploadSaveOption
	OpenUploadedFileFunc = netimpl.OpenUploadedFileFunc
	MultipartFormData    = netimpl.MultipartFormData
	LocalPortGenerator   = netimpl.LocalPortGenerator
)
