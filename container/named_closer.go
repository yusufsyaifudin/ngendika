package container

import (
	"io"
)

type Closer interface {
	io.Closer

	Name() string
}

type NamedCloser struct {
	name   string
	closer io.Closer
}

func (d *NamedCloser) Close() error {
	return d.closer.Close()
}

func (d *NamedCloser) Name() string {
	return d.name
}

var _ Closer = (*NamedCloser)(nil)

func NewNamedCloser(name string, closer io.Closer) *NamedCloser {
	return &NamedCloser{
		name:   name,
		closer: closer,
	}
}
