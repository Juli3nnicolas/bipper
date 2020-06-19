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

type UI interface {
	Run()
}

type TermDashUI struct {
	bip 			*bipper.Bipper
	bipFile			string
	endBipFile		string
	sectionFile		chan string
	currentSection 	chan string
	remainingTime 	chan time.Duration
	rawDocument 	chan string
}

func (o *TermDashUI) Init(bipFile, endBipFile string) {
	o.bipFile = bipFile
	o.endBipFile = endBipFile
	o.sectionFile = make(chan string)
	o.currentSection = make(chan string)
	o.remainingTime = make(chan time.Duration)
	o.rawDocument = make(chan string)
}

const (
	emptyCurrentSection string = "-"
	emptyRawDocument string = " "
	emptyRemainingTime time.Duration = time.Duration(0)
)

// redrawInterval is how often termdash redraws the screen.
const redrawInterval = 250 * time.Millisecond

// widgets holds the widgets used by this demo.
type widgets struct {
	currentSectionMessage	*segmentdisplay.SegmentDisplay
	openedFileMessage 		*textinput.TextInput
	blank					*text.Text
	rawDocument    			*text.Text
	remainingTime			*segmentdisplay.SegmentDisplay
}

// newWidgets creates all widgets used by this demo.
func (o *TermDashUI) newWidgets(c *container.Container) (*widgets, error) {
	openedFileMessage, err := newTextInput(o.sectionFile)
	if err != nil {
		return nil, err
	}

	currentSectionMessage, err := newSegmentDisplay(emptyCurrentSection, o.currentSection)
	if err != nil {
		return nil, err
	}

	blank, err := newTextLabel(" ")
	if err != nil {
		return nil, err
	}

	rawDocument, err := newRollText(o.rawDocument)
	if err != nil {
		return nil, err
	}

	remainingTime, err := newTimeSegmentDisplay(emptyRemainingTime.String(), o.remainingTime)
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

func (o *TermDashUI) Run() {
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
	w, err := o.newWidgets(c)
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

	// Poll UI messages
	go o.pollInput()

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		panic(err)
	}
}

func (o *TermDashUI) pollInput() {
	for {
		// This step is necessary in case no bipper has been set
		var currentSection, rawDocument, msg chan string
		var remainingTime chan time.Duration
		if o.bip != nil {
			currentSection = o.bip.Output.SectionName
			rawDocument = o.bip.Output.RawDoc
			msg = o.bip.Output.Msg
			remainingTime = o.bip.Output.Remaining
		}

		select {
		// Create a new bipper
		case file := <- o.sectionFile:
			if o.bip != nil {
				o.bip.Close()
			}
			o.bip = &bipper.Bipper{}
			
			err := o.bip.Init(o.bipFile, o.endBipFile, file)
			if err != nil {
				o.bip = nil
				o.currentSection <- emptyCurrentSection
				o.rawDocument <- emptyRawDocument
				o.remainingTime <- emptyRemainingTime
				break
			}

			go func() {
				o.bip.Bip()
				o.bip.Close()
			}()

			// Pass the messages to the UI
			case tmp := <- currentSection:
				o.currentSection <- tmp
			case tmp := <- rawDocument:
				o.rawDocument <- tmp
			case tmp := <- remainingTime:
				o.remainingTime <- tmp
			case <- msg:
		}
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
		textinput.Label("File path: ", cell.FgColor(cell.ColorWhite)),
		textinput.MaxWidthCells(20),
		textinput.PlaceHolder("click here"),
		textinput.PlaceHolderColor(cell.ColorWhite),
		textinput.FillColor(cell.ColorNumber(0)),
		textinput.OnSubmit(func(text string) error {
			updateText <- text
			return nil
		}),
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
			t.Reset()
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
