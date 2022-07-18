package cache

// statWriter implements io.Writer and keeps track of the written bytes.
type statWriter struct {
	written int64
}

func (s *statWriter) Write(p []byte) (int, error) {
	size := len(p)
	s.written += int64(size)

	return size, nil
}
