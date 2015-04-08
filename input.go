package rich

import (
	"fmt"
	"unicode"

	"github.com/nsf/termbox-go"
)

type EventHandler interface {
	HandleEvent(evt termbox.Event) bool
}

type StdInput struct {
	w *Widget
}

func NewStandardInput(w *Widget) *StdInput {
	return &StdInput{
		w: w,
	}
}

func (si *StdInput) HandleEvent(evt termbox.Event) bool {
	if evt.Type == termbox.EventKey {
		if si.w.CursorVisible() {
			switch evt.Key {
			case termbox.KeyArrowLeft:
				si.w.MoveCursor(-1)
			case termbox.KeyArrowRight:
				si.w.MoveCursor(1)
			case termbox.KeyHome:
				// Start of line
			case termbox.KeyEnd:
				// End of line
			case termbox.KeyArrowUp:
				// Up one line
			case termbox.KeyArrowDown:
				// Down one line
			case termbox.KeyPgup:
				// Up a bunch of lines
			case termbox.KeyPgdn:
				// Down a bunch of lines
			case termbox.KeyDelete:
				si.w.Delete(1)
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				si.w.Delete(-1)
			case termbox.KeyEnter:
				fmt.Fprintf(si.w, "\n")
			case termbox.KeySpace:
				fmt.Fprintf(si.w, " ")
			default:
				if unicode.IsPrint(evt.Ch) {
					fmt.Fprintf(si.w, string(evt.Ch))
				} else {
					// Dunno what this char is
				}
			}
			return true
		}
	}
	return false
}
