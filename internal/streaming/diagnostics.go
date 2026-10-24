package streaming

import "sync"

// stderrTail keeps diagnostics bounded even for malformed media files.
type stderrTail struct {
	mu   sync.Mutex
	data []byte
}

func (tail *stderrTail) Write(data []byte) (int, error) {
	tail.mu.Lock()
	defer tail.mu.Unlock()
	size := len(data)
	tail.data = append(tail.data, data...)
	if len(tail.data) > 8192 {
		tail.data = tail.data[len(tail.data)-8192:]
	}
	return size, nil
}
func (tail *stderrTail) String() string {
	tail.mu.Lock()
	defer tail.mu.Unlock()
	return string(tail.data)
}
func (stream *Stream) Diagnostics() string { return stream.diagnostics.String() }
