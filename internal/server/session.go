package server

// StackSession owns the per-request body handed to the stack solver.
// Close releases the session; it is not safe to call twice.
type StackSession struct {
	done   chan struct{}
	closed bool
}

// OpenStackSession starts a request session.
func OpenStackSession() *StackSession {
	return &StackSession{done: make(chan struct{})}
}

// Close ends the session. The done channel is closed exactly once.
func (s *StackSession) Close() {
	if s == nil || s.closed {
		return
	}
	close(s.done)
	s.closed = true
}

// Closed reports whether Close has run.
func (s *StackSession) Closed() bool {
	return s.closed
}
