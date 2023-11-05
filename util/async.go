package util

import "io"

type AsyncAction interface {
	Wait() error
	io.Closer
}
