package multidb

import "io"

type Closer interface {
	io.Closer

	Name() string
}

type namedCloser struct {
	name   string
	closer io.Closer
}

func (d *namedCloser) Close() error {
	return d.closer.Close()
}

func (d *namedCloser) Name() string {
	return d.name
}

var _ Closer = (*namedCloser)(nil)

func newNamedCloser(name string, closer io.Closer) *namedCloser {
	return &namedCloser{
		name:   name,
		closer: closer,
	}
}
