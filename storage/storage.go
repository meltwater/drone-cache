package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/go-kit/log"
	"github.com/meltwater/drone-cache/storage/backend"
)

const DefaultOperationTimeout = 30 * time.Minute

// Storage is a place that files can be written to and read from.
type Storage interface {
	// Get writes contents of the given object with given key from remote storage to io.Writer.
	Get(p string, w io.Writer) error

	// Put writes contents of io.Reader to remote storage at given key location.
	Put(p string, r io.Reader) error

	// Exists checks if object with given key exists in remote storage.
	Exists(p string) (bool, error)

	// List lists contents of the given directory by given key from remote storage.
	List(p string) ([]backend.FileEntry, error)

	// Delete deletes the object from remote storage.
	Delete(p string) error
}

// Default Storage implementation.
type storage struct {
	logger log.Logger

	b       backend.Backend
	timeout time.Duration
}

// New create a new default storage.
func New(l log.Logger, b backend.Backend, timeout time.Duration) Storage {
	return &storage{l, b, timeout}
}

// Get writes contents of the given object with given key from remote storage to io.Writer.
func (s *storage) Get(p string, w io.Writer) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	if err := s.b.Get(ctx, p, w); err != nil {
		return fmt.Errorf("storage backend get failure, %w", err)
	}

	return nil
}

// Put writes contents of io.Reader to remote storage at given key location.
func (s *storage) Put(p string, r io.Reader) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	if err := s.b.Put(ctx, p, r); err != nil {
		return fmt.Errorf("storage backend put failure, %w", err)
	}

	return nil
}

// Exists checks if object with given key exists in remote storage.
func (s *storage) Exists(p string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	ret, err := s.b.Exists(ctx, p)
	if err != nil {
		return ret, fmt.Errorf("storage backend exists failure, %w", err)
	}

	return ret, nil
}

// List lists contents of the given directory by given key from remote storage.
func (s *storage) List(p string) ([]backend.FileEntry, error) {
	// Implement me!
	// Make sure consumer utilizes context.
	return []backend.FileEntry{}, nil
}

// Delete deletes the object from remote storage.
func (s *storage) Delete(p string) error {
	// Implement me!
	// Make sure consumer utilizes context.
	return nil
}
