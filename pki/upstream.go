package pki

import (
	"rahnit-rmm/config"
)

const upstremCertFile = "upstream.crt"

var Upstream = &storedCertificate{
	path:          config.GetFilePath(upstremCertFile),
	allowOverride: false,
}
