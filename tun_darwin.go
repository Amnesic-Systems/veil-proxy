package proxy

import (
	"os"
)

// nitriding-proxy does not support macOS but we can at least make it compile by
// implementing the following functions.
const err = "not implemented on darwin"

func SetupTunAsProxy() error {
	panic(err)
}

func SetupTunAsEnclave() error {
	panic(err)
}

func CreateTun() (*os.File, error) {
	panic(err)
}
