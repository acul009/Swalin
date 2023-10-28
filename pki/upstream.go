package pki

const upstremCertFile = "upstream.crt"

var Upstream = &storedCertificate{
	filename:      upstremCertFile,
	allowOverride: false,
}
