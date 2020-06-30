package ui

import (
	"image"
	"sync"

	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
)

type Pauser struct {
	paused bool
	ch     chan bool
	key    keyboard.Key

	// mu protects the widget.
	mu sync.Mutex
}

func NewPauser(k keyboard.Key, ch chan bool) *Pauser {
	p := &Pauser{}
	p.key = k
	p.ch = ch

	return p
}

// PauseKeyDown receives a toggle value whenever the pause key is pressed
func (o *Pauser) PauseKeyDown() chan bool {
	return o.ch
}

func (o *Pauser) pause() error {
	o.ch <- o.paused
	return nil
}

////////////////////////////////////////////////////////////////////////////
//
// 					W I D G E T   I N T E R F A C E
//
////////////////////////////////////////////////////////////////////////////

// When the infrastructure calls Draw(), the widget must block on the call
// until it finishes drawing onto the provided canvas. When given the
// canvas, the widget must first determine its size by calling
// Canvas.Size(), then limit all its drawing to this area.
//
// The widget must not assume that the size of the canvas or its content
// remains the same between calls.
//
// The argument meta is guaranteed to be valid (i.e. non-nil).
// NOTE: THIS WIDGET DOESN'T DRAW ANYTHING
func (o *Pauser) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	return nil
}

// activated asserts whether the keyboard event activated the button.
func (b *Pauser) keyActivated(k *terminalapi.Keyboard) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if k.Key == b.key {
		b.paused = !b.paused
		return true
	}
	return false
}

// Keyboard processes keyboard events, acts as a button press on the configured
// Key.
//
// Implements widgetapi.Widget.Keyboard.
func (o *Pauser) Keyboard(k *terminalapi.Keyboard) error {
	if o.keyActivated(k) {
		// Mutex must be released when calling the callback.
		// Users might call container methods from the callback like the
		// Container.Update, see #205.
		return o.pause()
	}
	return nil
}

// Mouse is called when the widget is focused on the dashboard and a mouse
// event happens on its canvas. Only called if the widget registered for mouse
// events.
func (o *Pauser) Mouse(m *terminalapi.Mouse) error {
	return nil
}

// Options returns registration options for the widget.
// This is how the widget indicates to the infrastructure whether it is
// interested in keyboard or mouse shortcuts, what is its minimum canvas
// size, etc.
//
// Most widgets will return statically compiled options (minimum and
// maximum size, etc.). If the returned options depend on the runtime state
// of the widget (e.g. the user data provided to the widget), the widget
// must protect against a case where the infrastructure calls the Draw
// method with a canvas that doesn't meet the requested options. This is
// because the data in the widget might change between calls to Options and
// Draw.
func (o *Pauser) Options() widgetapi.Options {
	return widgetapi.Options{
		MinimumSize:  image.Point{1, 1},
		MaximumSize:  image.Point{1, 1},
		WantKeyboard: widgetapi.KeyScopeGlobal,
	}
}
