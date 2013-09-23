package main

import (
	"fmt"
)

// A difference in the screen appearance.  Can be applied to a Telnet
// interface to be sent over the network.
type ScreenDelta struct {
	x, y int
	data []byte
}

func (d ScreenDelta) Apply(t Telnet) {
	t.GoTo(uint16(d.x + 1), uint16(d.y + 1))
	t.Write(d.data)
}

func (d ScreenDelta) String() string {
	return fmt.Sprintf("x: %d y: %d data: %s", d.x, d.y, string(d.data))
}

// An abstract screen that is viewed by the player.
type Screen interface {
	Flip()
	GetDelta() []ScreenDelta
	GoTo(x, y int)
	Put(b ...byte)
	Write(b []byte)
	GetSize() (int, int)
}

type screen struct {
	width, height	int
	cx, cy		int
	currentBuffer	int
	buffer		[][]byte
	delta		[]byte
}

// MakeScreen creates a new screen of a specified width and height.
func MakeScreen(width, height int) Screen {
	s := &screen{
		width: width,
		height: height,
		cx: 0,
		cy: 0,
		currentBuffer: 0}

	s.buffer = make([][]byte, 2)
	for i := range s.buffer {
		s.buffer[i] = make([]byte, width * height)
	}

	s.delta = make([]byte, width * height)
	return s
}

// Flip switches to a new screen buffer.
func (s *screen) Flip() {
	s.currentBuffer = 1 - s.currentBuffer
}

// GetDelta returns the difference between the current and last screen.
func (s screen) GetDelta() []ScreenDelta {
	cur := s.getCurrentBuffer()
	last := s.buffer[1 - s.currentBuffer]

	for i := range s.delta {
		s.delta[i] = cur[i] - last[i]
	}

	delta := make([]ScreenDelta, 0, 30)

	for r := 0; r < s.height; r++ {
		i, j := 0, 0
		for i < s.width {
			for ; j < s.width; j++ {
				if s.delta[r * s.width + j] != 0 {
					break
				}
			}

			i = j

			for ; j < s.width; j++ {
				if s.delta[r * s.width + j] == 0 {
					break
				}
			}

			if i == j {
				break
			}

			delta = append(delta, ScreenDelta{i, r, cur[r * s.width + i:r * s.width + j]})
		}
	}

	return delta
}

// GoTo places the write cursor at the specified position.
func (s *screen) GoTo(x, y int) {
	s.cx = x
	s.cy = y
}

// getCurrentBuffer returns the current active buffer.
func (s screen) getCurrentBuffer() []byte {
	return s.buffer[s.currentBuffer]
}

// getCursorIndex returns the current linear cursor position.
func (s screen) getCursorIndex() int {
	return s.cy * s.width + s.cx
}

// Put writes a list of bytes to the screen.
func (s *screen) Put(b ...byte) {
	s.Write(b)
}

// Write writes an array of bytes to the screen.
func (s *screen) Write(b []byte) {
	n := Mini(len(b), s.width - s.cx)
	for i := 0; i < n; i++ {
		s.getCurrentBuffer()[s.getCursorIndex()] = b[i]
		s.cx++
	}
}

// GetSize returns the width and height of the screen.
func (s screen) GetSize() (int, int) {
	return s.width, s.height
}
