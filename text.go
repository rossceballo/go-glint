package dynamiccli

import (
	"sync"
)

// TextElements is an Element that renders text.
type TextElement struct {
	mu   sync.Mutex
	text string
}

// Text creates a TextElement for static text. The text here will be word
// wrapped automatically based on the width of the terminal.
func Text(v string) *TextElement {
	return &TextElement{
		text: v,
	}
}

// Update updates the text element. This is safe to call while this is being
// rendered.
func (el *TextElement) Update(text string) {
	el.mu.Lock()
	defer el.mu.Unlock()
	el.text = text
}

func (el *TextElement) Render(width uint) string {
	el.mu.Lock()
	defer el.mu.Unlock()
	return el.text
}
