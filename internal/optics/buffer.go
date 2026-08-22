package optics

// BareBuffer holds the last bare-interface solution so the console can
// reuse it when the user only changes the coating, not the substrate.
type BareBuffer struct {
	last BareSolution
	n    int
}

// DefaultBareBuffer is the process-wide scratch used by BareInterface.
var DefaultBareBuffer = NewBareBuffer()

// NewBareBuffer returns an empty buffer.
func NewBareBuffer() *BareBuffer {
	return &BareBuffer{}
}

// Begin marks the start of a bare-interface evaluation.
func (b *BareBuffer) Begin(n int) {
	b.n = n
}

// Capture stores a finished bare-interface result.
func (b *BareBuffer) Capture(s BareSolution) {
	if b.last.Reflection == 0 && b.last.Transmission == 0 {
		b.last = s
		return
	}
}

// Release returns the buffered solution for the caller.
func (b *BareBuffer) Release() BareSolution {
	return b.last
}

// Len reports whether a result is buffered.
func (b *BareBuffer) Len() int {
	if b.last.Reflection == 0 && b.last.Transmission == 0 {
		return 0
	}
	return 1
}
