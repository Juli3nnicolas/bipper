package ui

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/Juli3nnicolas/bipper/pkg/bipper"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/segmentdisplay"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
)

// redrawInterval is how often termdash redraws the screen.
const redrawInterval = 250 * time.Millisecond

// widgets holds the widgets used by this demo.
type widgets struct {
	currentSectionMessage	*segmentdisplay.SegmentDisplay
	openedFileMessage 		*text.Text
	blank					*text.Text
	rawDocument    				*text.Text
	remainingTime			*segmentdisplay.SegmentDisplay
}

// newWidgets creates all widgets used by this demo.
func newWidgets(input bipper.BipperOutput, c *container.Container) (*widgets, error) {
	openedFileMessage, err := newTextLabel("Example.yaml")
	if err != nil {
		return nil, err
	}

	currentSectionMessage, err := newSegmentDisplay("Unknown", input.SectionName)
	if err != nil {
		return nil, err
	}

	blank, err := newTextLabel(" ")
	if err != nil {
		return nil, err
	}

	rawDocument, err := newRollText(input.RawDoc)
	if err != nil {
		return nil, err
	}

	remainingTime, err := newTimeSegmentDisplay("0", input.Remaining)
	if err != nil {
		return nil, err
	}

	return &widgets{
		openedFileMessage: openedFileMessage,
		currentSectionMessage: currentSectionMessage,
		blank: blank,
		rawDocument: rawDocument,
		remainingTime: remainingTime,
	}, nil
}

// gridLayout prepares container options that represent the desired screen layout.
// This function demonstrates the use of the grid builder.
// gridLayout() and contLayout() demonstrate the two available layout APIs and
// both produce equivalent layouts for layoutType layoutAll.
func gridLayout(w *widgets) ([]container.Option, error) {

	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(5, grid.Widget(w.openedFileMessage,
			container.Border(linestyle.None),
		),),
		grid.RowHeightPerc(25, grid.Widget(w.currentSectionMessage,
			container.Border(linestyle.None),
		),),
		grid.RowHeightPerc(5, grid.Widget(w.blank,
			container.Border(linestyle.None),
		),),
		grid.RowHeightPerc(65,
				grid.ColWidthPerc(20,
					grid.Widget(w.rawDocument,
						container.Border(linestyle.None),
					),
				),
				grid.ColWidthPerc(80,
					grid.Widget(w.remainingTime,
						container.Border(linestyle.None),
					),
				),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return gridOpts, nil
}

// rootID is the ID assigned to the root container.
const rootID = "root"

// Terminal implementations
const (
	termboxTerminal = "termbox"
	tcellTerminal   = "tcell"
)

func Tui(input bipper.BipperOutput) {
	terminalPtr := flag.String("terminal",
		"termbox",
		"The terminal implementation to use. Available implementations are 'termbox' and 'tcell' (default = termbox).")
	flag.Parse()

	var t terminalapi.Terminal
	var err error
	switch terminal := *terminalPtr; terminal {
	case termboxTerminal:
		t, err = termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	case tcellTerminal:
		t, err = tcell.New(tcell.ColorMode(terminalapi.ColorMode256))
	default:
		log.Fatalf("Unknown terminal implementation '%s' specified. Please choose between 'termbox' and 'tcell'.", terminal)
		return
	}

	if err != nil {
		panic(err)
	}
	defer t.Close()

	c, err := container.New(t, container.ID(rootID))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	w, err := newWidgets(input, c)
	if err != nil {
		panic(err)
	}

	gridOpts, err := gridLayout(w) // equivalent to contLayout(w)
	if err != nil {
		panic(err)
	}

	if err := c.Update(rootID, gridOpts...); err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
			cancel()
		}
	}

	// Poll unused output
	go func (input bipper.BipperOutput) {
		for {
			select {
			case <- input.Msg:
			}
		}
	}(input)

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		panic(err)
	}
}

// periodic executes the provided closure periodically every interval.
// Exits when the context expires.
func periodic(ctx context.Context, interval time.Duration, fn func() error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := fn(); err != nil {
				panic(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// textState creates a rotated state for the text we are displaying.
func textState(text string, capacity, step int) []rune {
	if capacity == 0 {
		return nil
	}

	var state []rune
	for i := 0; i < capacity; i++ {
		state = append(state, ' ')
	}
	state = append(state, []rune(text)...)
	step = step % len(state)
	return rotateRunes(state, step)
}

// newTextInput creates a new TextInput field that changes the text on the
// SegmentDisplay.
func newTextInput(updateText chan<- string) (*textinput.TextInput, error) {
	input, err := textinput.New(
		textinput.Label("Change text to: ", cell.FgColor(cell.ColorBlue)),
		textinput.MaxWidthCells(20),
		textinput.PlaceHolder("enter any text"),
		textinput.OnSubmit(func(text string) error {
			updateText <- text
			return nil
		}),
		textinput.ClearOnSubmit(),
	)
	if err != nil {
		return nil, err
	}
	return input, err
}

func newTextLabel(msg string) (*text.Text, error) {
	txt, err := text.New()
	txt.Write(msg, text.WriteCellOpts(cell.FgColor(cell.ColorWhite)))
	if err != nil {
		return nil, err
	}
	return txt, err
}

func updateChunks(sd *segmentdisplay.SegmentDisplay, text string, color cell.Color) {
	var chunks []*segmentdisplay.TextChunk
	chunks = append(chunks, segmentdisplay.NewChunk(
		text,
		segmentdisplay.WriteCellOpts(cell.FgColor(color)),
	))
	if err := sd.Write(chunks); err != nil {
		panic(err)
	}
}

// newSegmentDisplay creates a new SegmentDisplay that initially shows the
// Termdash name. Shows any text that is sent over the channel.
func newSegmentDisplay(initMsg string, textChan chan string) (*segmentdisplay.SegmentDisplay, error) {
	sd, err := segmentdisplay.New()
	if err != nil {
		return nil, err
	}

	/*colors := []cell.Color{
		cell.ColorBlue,
		cell.ColorRed,
		cell.ColorYellow,
		cell.ColorBlue,
		cell.ColorGreen,
		cell.ColorRed,
		cell.ColorGreen,
		cell.ColorRed,
	}*/

	text := strings.Repeat(" ", 9) + initMsg
	updateChunks(sd, text, cell.ColorYellow)

	go func (ch chan string) {
		for {
			newTxt := <- ch
			updateChunks(sd, newTxt, cell.ColorYellow)
		}
	}(textChan)

	return sd, nil
}

// newTimeSegmentDisplay creates a new SegmentDisplay that initially shows the
// Termdash name. Shows any text that is sent over the channel.
func newTimeSegmentDisplay(initMsg string, timeChan chan time.Duration) (*segmentdisplay.SegmentDisplay, error) {
	sd, err := segmentdisplay.New()
	if err != nil {
		return nil, err
	}

	/*colors := []cell.Color{
		cell.ColorBlue,
		cell.ColorRed,
		cell.ColorYellow,
		cell.ColorBlue,
		cell.ColorGreen,
		cell.ColorRed,
		cell.ColorGreen,
		cell.ColorRed,
	}*/

	text := initMsg
	updateChunks(sd, text, cell.ColorGreen)

	go func (ch chan time.Duration) {
		for {
			t := <- ch
			color := cell.ColorGreen
			if t.Seconds() <= 3.0 {
				color = cell.ColorRed
			}

			updateChunks(sd, t.String(), color)
		}
	}(timeChan)

	return sd, nil
}

// newRollText creates a new Text widget that displays rolling text.
func newRollText(ch chan string) (*text.Text, error) {
	t, err := text.New(text.RollContent())
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			txt := <- ch
			if err := t.Write(txt, text.WriteCellOpts(cell.FgColor(cell.ColorWhite))); err != nil {
				panic(err)
			}
		}
	}()

	return t, nil
}

// setLayout sets the specified layout.
func setLayout(c *container.Container, w *widgets) error {
	gridOpts, err := gridLayout(w)
	if err != nil {
		return err
	}
	return c.Update(rootID, gridOpts...)
}

// layoutButtons are buttons that change the layout.
type layoutButtons struct {
	allB  *button.Button
	textB *button.Button
	spB   *button.Button
	lcB   *button.Button
}

// rotateFloats returns a new slice with inputs rotated by step.
// I.e. for a step of one:
//   inputs[0] -> inputs[len(inputs)-1]
//   inputs[1] -> inputs[0]
// And so on.
func rotateFloats(inputs []float64, step int) []float64 {
	return append(inputs[step:], inputs[:step]...)
}

// rotateRunes returns a new slice with inputs rotated by step.
// I.e. for a step of one:
//   inputs[0] -> inputs[len(inputs)-1]
//   inputs[1] -> inputs[0]
// And so on.
func rotateRunes(inputs []rune, step int) []rune {
	return append(inputs[step:], inputs[:step]...)
}
