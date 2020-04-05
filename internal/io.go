package internal

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// CloseWithErrLogf is making sure we log every error, even those from best effort tiny closers.
func CloseWithErrLogf(logger log.Logger, closer io.Closer, format string, a ...interface{}) {
	err := close(closer)
	if err == nil {
		return
	}

	if logger == nil {
		logger = log.NewLogfmtLogger(os.Stderr)
	}

	level.Warn(logger).Log("msg", "detected close error", "err", fmt.Errorf(format+", %w", append(a, err)...))
}

// CloseWithErrCapturef runs function and on error return error by argument including the given error..
func CloseWithErrCapturef(err *error, closer io.Closer, format string, a ...interface{}) {
	if err != nil {
		cErr := close(closer)
		if cErr == nil {
			return
		}

		mErr := MultiError{}
		mErr.Add(*err)
		mErr.Add(fmt.Errorf(format+", %w", append(a, cErr)...))
		*err = mErr.Err()

		return
	}

	cErr := close(closer)
	if cErr == nil {
		return
	}

	*err = cErr
}

func close(closer io.Closer) error {
	err := closer.Close()
	if err == nil {
		return nil
	}

	if errors.Is(err, os.ErrClosed) {
		return nil
	}

	return err
}
