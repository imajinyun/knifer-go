package vnet

import netimpl "github.com/imajinyun/go-knifer/internal/net"

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
	TLSConfigBuilder   = netimpl.TLSConfigBuilder
	UploadSetting      = netimpl.UploadSetting
	UploadSaveOption   = netimpl.UploadSaveOption
	MultipartFormData  = netimpl.MultipartFormData
	LocalPortGenerator = netimpl.LocalPortGenerator
)
