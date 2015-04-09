package rich

import (
	"bytes"
	"sync"
	"time"

	"github.com/gizak/termui"
)

const BlinkRate = time.Millisecond * 600

func chanWait(d time.Duration) chan struct{} {
	c := make(chan struct{})
	go func() {
		time.Sleep(d)
		c <- struct{}{}
	}()
	return c
}

type cell struct {
	ch     rune
	fg, bg termui.Attribute
}

func (c cell) point(x, y int) termui.Point {
	return termui.Point{
		Ch: c.ch,
		Fg: c.fg,
		Bg: c.bg,
		X:  x,
		Y:  y,
	}
}

type lineInfo struct {
	pos, length int
}

type Widget struct {
	termui.Block
	WriteFg, WriteBg termui.Attribute

	MultiLine bool

	wrap bool

	cursorMut        sync.Mutex
	cursorEnabled    bool
	cursorBlinkState bool
	cursorPos        int
	cursorCancel     chan struct{}

	scrollX, scrollY int

	dirtyHandlers []func()

	writeMut sync.Mutex
	contents []cell
	lines    []lineInfo
}

func New() *Widget {
	return &Widget{
		Block:         *termui.NewBlock(),
		MultiLine:     true,
		dirtyHandlers: []func(){},
		cursorCancel:  make(chan struct{}),
	}
}

func (w *Widget) AddDirtyHandler(h func()) {
	w.dirtyHandlers = append(w.dirtyHandlers, h)
}

func (w Widget) dirty() {
	for _, h := range w.dirtyHandlers {
		h()
	}
}

func (w Widget) Buffer() []termui.Point {
	points := w.Block.Buffer()
	innerX, innerY, innerWidth, innerHeight := w.InnerBounds()
	col := innerX
	line := innerY
	overflow := false
	for i, c := range w.contents {
		p := c.point(col, line)
		if c.ch == '\n' {
			p.Ch = 0
			line++
			if line >= innerY+innerHeight {
				overflow = true
				break
			}
			col = innerX
		}
		if w.cursorEnabled && w.cursorBlinkState && w.cursorPos == i {
			p.Fg ^= termui.AttrReverse
			p.Bg ^= termui.AttrReverse
			if p.Ch == 0 {
				p.Ch = ' '
				col-- // Column will be incremented when the point is output.
			}
		}
		if p.Ch != 0 {
			points = append(points, p)
			col++
			if col >= innerX+innerWidth {
				col = innerX
				line++
				if line >= innerY+innerHeight {
					overflow = true
					break
				}
			}
		}
	}
	// Special case, end of text and blink is on.
	if !overflow && w.cursorEnabled && w.cursorBlinkState && w.cursorPos == len(w.contents) {
		points = append(points, termui.Point{
			Ch: ' ',
			Fg: termui.AttrReverse,
			Bg: termui.AttrReverse,
			X:  col,
			Y:  line,
		})
	}
	return points
}

func (w *Widget) Write(p []byte) (n int, err error) {
	l := 0
	insert := []cell{}
	for _, ch := range string(p) {
		if ch == '\n' && !w.MultiLine {
			continue
		}
		insert = append(insert, cell{ch, w.WriteFg, w.WriteBg})
		l++
	}
	tail := append([]cell{}, w.contents[w.cursorPos:]...)
	w.contents = append(append(w.contents[:w.cursorPos], insert...), tail...)
	w.MoveCursor(l)
	w.dirty()
	return len(p), nil
}

func (w *Widget) CursorShow() {
	if w.cursorEnabled {
		return
	}
	w.cursorMut.Lock()
	defer w.cursorMut.Unlock()
	w.cursorEnabled = true
	w.cursorBlinkState = false
	go w.cursorBlink()
}

func (w *Widget) CursorHide() {
	if !w.cursorEnabled {
		return
	}
	w.cursorMut.Lock()
	defer w.cursorMut.Unlock()
	w.cursorCancel <- struct{}{}
	w.cursorEnabled = false
}

func (w *Widget) CursorVisible() bool {
	return w.cursorEnabled
}

func (w *Widget) SetCursorLoc(x, y int) bool {
	return false
}

func (w *Widget) SetCursorPos(pos int) bool {
	w.cursorPos = pos
	if w.cursorPos < 0 {
		w.cursorPos = 0
	}
	if l := len(w.contents); w.cursorPos > l {
		w.cursorPos = l
	}
	if w.cursorEnabled {
		w.CursorHide()
		w.CursorShow()
	}
	return true
}

func (w *Widget) CursorPos() int {
	return w.cursorPos
}

func (w *Widget) MoveCursor(n int) bool {
	return w.SetCursorPos(w.cursorPos + n)
}

func (w *Widget) cursorBlink() {
	w.cursorBlinkState = !w.cursorBlinkState
	w.dirty()
	select {
	case <-chanWait(BlinkRate):
		w.cursorBlink()
	case <-w.cursorCancel:
		// Let the function exit.
	}
}

func (w *Widget) WrapOn() {
	w.SetWrap(true)
}

func (w *Widget) WrapOff() {
	w.SetWrap(false)
}

func (w *Widget) SetWrap(wrap bool) {
	if wrap != w.wrap {
		defer w.dirty()
	}
	w.wrap = wrap
}

func (w *Widget) Delete(n int) {
	switch {
	case n > 0:
		if l := len(w.contents); w.cursorPos+n > l {
			n = l - w.cursorPos
		}
		w.contents = append(w.contents[:w.cursorPos], w.contents[w.cursorPos+n:]...)
	case n < 0:
		if n < -w.cursorPos {
			n = -w.cursorPos
		}
		w.contents = append(w.contents[:w.cursorPos+n], w.contents[w.cursorPos:]...)
		w.MoveCursor(n)
	}
	w.dirty()
}

func (w *Widget) String() string {
	buf := bytes.Buffer{}
	for _, c := range w.contents {
		buf.WriteRune(c.ch)
	}
	return buf.String()
}
