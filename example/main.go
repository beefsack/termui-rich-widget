package main

import (
	"log"

	"github.com/beefsack/termui-rich-widget"
	"github.com/gizak/termui"
	"github.com/nsf/termbox-go"
)

func main() {
	if err := termui.Init(); err != nil {
		log.Fatalf("failed to initialise UI: %s", err)
	}
	defer termui.Close()

	r := rich.New()

	// rich can update it's own content (handling input and cursor blinking) so
	// registering a dirty handler lets us render when it changes.
	r.AddDirtyHandler(func() {
		termui.Render(r)
	})

	r.CursorShow()

	handler := rich.NewStandardInput(r)
	for {
		evt := termbox.PollEvent()
		if evt.Type == termbox.EventKey && evt.Key == termbox.KeyEsc {
			log.Println(r)
			break
		}
		handler.HandleEvent(evt)
	}
}
