package internal

import (
	"bytes"
	"fmt"
	"sync"
)

// NOTICE: Modified version of https://github.com/prometheus/prometheus/blob/master/tsdb/errors/errors.go.

// The MultiError type implements the error interface, and contains the Errors used to construct it.
type MultiError struct {
	mu   sync.Mutex
	errs []error
}

// Returns a concatenated string of the contained errors.
func (me *MultiError) Error() string {
	var buf bytes.Buffer

	me.mu.Lock()
	defer me.mu.Unlock()

	if len(me.errs) > 1 {
		fmt.Fprintf(&buf, "%d errors: ", len(me.errs))
	}

	for i, err := range me.errs {
		if i != 0 {
			buf.WriteString(";\n")
		}

		buf.WriteString(err.Error())
	}

	return buf.String()
}

// Add adds the error to the error list if it is not nil.
func (me *MultiError) Add(err error) {
	if err == nil {
		return
	}

	me.mu.Lock()
	defer me.mu.Unlock()

	me.errs = append(me.errs, err)
}

// Err returns the error list as an error or nil if it is empty.
func (me *MultiError) Err() error {
	me.mu.Lock()
	defer me.mu.Unlock()

	if len(me.errs) == 0 {
		return nil
	}

	return me
}
