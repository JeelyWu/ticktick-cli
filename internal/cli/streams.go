package cli

import "io"

type Streams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}
