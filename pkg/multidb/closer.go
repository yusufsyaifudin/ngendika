package multidb

import (
	"fmt"
	"io"
)

type Closer interface {
	io.Closer

	String() string
}

type namedCloser struct {
	name   string
	closer io.Closer
}

func (d *namedCloser) Close() error {
	err := d.closer.Close()
	if err != nil {
		err = fmt.Errorf("(%s) %w", d.name, err)
	}

	return err
}

func (d *namedCloser) String() string {
	return d.name
}

var _ Closer = (*namedCloser)(nil)

func newNamedCloser(name string, closer io.Closer) *namedCloser {
	return &namedCloser{
		name:   name,
		closer: closer,
	}
}
