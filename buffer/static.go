// Copyright (c) 2018 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buffer

// Static is a fixed-capacity buffer, for wrapping a memory-mapped region.  The
// default value is a zero-capacity buffer.
type Static struct {
	buf []byte
	max int
}

// MakeStatic buffer with len(b) as its capacity.  It is initially empty (b is
// truncated).
//
// This function can be used in field initializer expressions.  The initialized
// field must not be copied.
func MakeStatic(b []byte) Static {
	return Static{b[:0], len(b)}
}

// NewStatic buffer with len(b) as its capacity.  It is initially empty (b is
// truncated).
func NewStatic(b []byte) *Static {
	return &Static{b[:0], len(b)}
}

// Capacity of the static buffer.
func (s *Static) Cap() int {
	return s.max
}

// Bytes doesn't panic.
func (s *Static) Bytes() []byte {
	return s.buf
}

// PutByte panics with ErrStaticSize if the buffer is already full.
func (s *Static) PutByte(value byte) {
	offset := len(s.buf)
	if offset >= s.max {
		panic(ErrStaticSize)
	}
	s.buf = s.buf[:offset+1]
	s.buf[offset] = value
}

// Extend panics with ErrStaticSize if n bytes cannot be appended to the
// buffer.
func (s *Static) Extend(n int) []byte {
	offset := len(s.buf)
	size := offset + n
	if size > s.max {
		panic(ErrStaticSize)
	}
	s.buf = s.buf[:size]
	return s.buf[offset:]
}

// ResizeBytes panics with ErrStaticSize if n is larger than maximum buffer
// size.
func (s *Static) ResizeBytes(n int) []byte {
	if n > s.max {
		panic(ErrStaticSize)
	}
	s.buf = s.buf[:n]
	return s.buf
}