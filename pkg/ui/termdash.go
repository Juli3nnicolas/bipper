package ui

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/barchart"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/segmentdisplay"
	"github.com/mum4k/termdash/widgets/sparkline"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
)

// redrawInterval is how often termdash redraws the screen.
const redrawInterval = 250 * time.Millisecond

// widgets holds the widgets used by this demo.
type widgets struct {
	segDist  				*segmentdisplay.SegmentDisplay
	openedFileMessage 		*text.Text
	currentSectionMessage 	*text.Text
	input    				*textinput.TextInput
	rollT    				*text.Text
	spGreen  				*sparkline.SparkLine
	spRed    				*sparkline.SparkLine
	gauge    				*gauge.Gauge
	heartLC  				*linechart.LineChart
	barChart 				*barchart.BarChart
	donut    				*donut.Donut
	leftB    				*button.Button
	rightB   				*button.Button
	sineLC   				*linechart.LineChart
	buttons 				*layoutButtons
}

// newWidgets creates all widgets used by this demo.
func newWidgets(ctx context.Context, c *container.Container) (*widgets, error) {
	updateText := make(chan string)
	sd, err := newSegmentDisplay(ctx, updateText)
	if err != nil {
		return nil, err
	}

	input, err := newTextInput(updateText)
	if err != nil {
		return nil, err
	}

	openedFileMessage, err := newTextLabel("Example.yaml")
	if err != nil {
		return nil, err
	}

	currentSectionMessage, err := newTextLabel("Burpees")
	if err != nil {
		return nil, err
	}

	rollT, err := newRollText(ctx)
	if err != nil {
		return nil, err
	}
	spGreen, spRed, err := newSparkLines(ctx)
	if err != nil {
		return nil, err
	}
	g, err := newGauge(ctx)
	if err != nil {
		return nil, err
	}

	bc, err := newBarChart(ctx)
	if err != nil {
		return nil, err
	}

	don, err := newDonut(ctx)
	if err != nil {
		return nil, err
	}

	return &widgets{
		openedFileMessage: openedFileMessage,
		currentSectionMessage: currentSectionMessage,
		segDist:  sd,
		input:    input,
		rollT:    rollT,
		spGreen:  spGreen,
		spRed:    spRed,
		gauge:    g,
		barChart: bc,
		donut:    don,
	}, nil
}

// gridLayout prepares container options that represent the desired screen layout.
// This function demonstrates the use of the grid builder.
// gridLayout() and contLayout() demonstrate the two available layout APIs and
// both produce equivalent layouts for layoutType layoutAll.
func gridLayout(w *widgets) ([]container.Option, error) {

	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(10, grid.Widget(w.openedFileMessage,
			container.Border(linestyle.Light),
			container.BorderTitle("opened file"),
		),),
		grid.RowHeightPerc(10, grid.Widget(w.segDist,
			container.Border(linestyle.None),
		),),
		grid.RowHeightPerc(80,
				grid.ColWidthPerc(20,
					grid.Widget(w.rollT,
						container.Border(linestyle.None),
					),
				),
				grid.ColWidthPerc(80,
					grid.Widget(w.donut,
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

// contLayout prepares container options that represent the desired screen layout.
// This function demonstrates the direct use of the container API.
// gridLayout() and contLayout() demonstrate the two available layout APIs and
// both produce equivalent layouts for layoutType layoutAll.
// contLayout only produces layoutAll.
func contLayout(w *widgets) ([]container.Option, error) {
	buttonRow := []container.Option{
		container.SplitVertical(
			container.Left(
				container.SplitVertical(
					container.Left(
						container.PlaceWidget(w.buttons.allB),
					),
					container.Right(
						container.PlaceWidget(w.buttons.textB),
					),
				),
			),
			container.Right(
				container.SplitVertical(
					container.Left(
						container.PlaceWidget(w.buttons.spB),
					),
					container.Right(
						container.PlaceWidget(w.buttons.lcB),
					),
				),
			),
		),
	}

	textAndSparks := []container.Option{
		container.SplitVertical(
			container.Left(
				container.Border(linestyle.Light),
				container.BorderTitle("A rolling text"),
				container.PlaceWidget(w.rollT),
			),
			container.Right(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("Green SparkLine"),
						container.PlaceWidget(w.spGreen),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle("Red SparkLine"),
						container.PlaceWidget(w.spRed),
					),
				),
			),
		),
	}

	segmentTextInputSparks := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("Press Esc to quit"),
				container.PlaceWidget(w.segDist),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.SplitHorizontal(
							container.Top(
								container.PlaceWidget(w.input),
							),
							container.Bottom(buttonRow...),
						),
					),
					container.Bottom(textAndSparks...),
					container.SplitPercent(40),
				),
			),
			container.SplitPercent(50),
		),
	}

	gaugeAndHeartbeat := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("A Gauge"),
				container.BorderColor(cell.ColorNumber(39)),
				container.PlaceWidget(w.gauge),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.BorderTitle("A LineChart"),
				container.PlaceWidget(w.heartLC),
			),
			container.SplitPercent(20),
		),
	}

	leftSide := []container.Option{
		container.SplitHorizontal(
			container.Top(segmentTextInputSparks...),
			container.Bottom(gaugeAndHeartbeat...),
			container.SplitPercent(50),
		),
	}

	lcAndButtons := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("Multiple series"),
				container.BorderTitleAlignRight(),
				container.PlaceWidget(w.sineLC),
			),
			container.Bottom(
				container.SplitVertical(
					container.Left(
						container.PlaceWidget(w.leftB),
						container.AlignHorizontal(align.HorizontalRight),
						container.PaddingRight(1),
					),
					container.Right(
						container.PlaceWidget(w.rightB),
						container.AlignHorizontal(align.HorizontalLeft),
						container.PaddingLeft(1),
					),
				),
			),
			container.SplitPercent(80),
		),
	}

	rightSide := []container.Option{
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("BarChart"),
				container.PlaceWidget(w.barChart),
				container.BorderTitleAlignRight(),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("A Donut"),
						container.BorderTitleAlignRight(),
						container.PlaceWidget(w.donut),
					),
					container.Bottom(lcAndButtons...),
					container.SplitPercent(30),
				),
			),
			container.SplitPercent(30),
		),
	}

	return []container.Option{
		container.SplitVertical(
			container.Left(leftSide...),
			container.Right(rightSide...),
			container.SplitPercent(70),
		),
	}, nil
}

// rootID is the ID assigned to the root container.
const rootID = "root"

// Terminal implementations
const (
	termboxTerminal = "termbox"
	tcellTerminal   = "tcell"
)

func tui() {
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
	w, err := newWidgets(ctx, c)
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

// newSegmentDisplay creates a new SegmentDisplay that initially shows the
// Termdash name. Shows any text that is sent over the channel.
func newSegmentDisplay(ctx context.Context, updateText <-chan string) (*segmentdisplay.SegmentDisplay, error) {
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

	text := strings.Repeat(" ", 9) + "Termdash"

	var chunks []*segmentdisplay.TextChunk
	chunks = append(chunks, segmentdisplay.NewChunk(
		text,
		segmentdisplay.WriteCellOpts(cell.FgColor(cell.ColorGreen)),
	))
	if err := sd.Write(chunks); err != nil {
		panic(err)
	}

	return sd, nil
}

// newRollText creates a new Text widget that displays rolling text.
func newRollText(ctx context.Context) (*text.Text, error) {
	t, err := text.New(text.RollContent())
	if err != nil {
		return nil, err
	}

	i := 0
	go periodic(ctx, 1*time.Second, func() error {
		if err := t.Write(fmt.Sprintf("Writing line %d.\n", i), text.WriteCellOpts(cell.FgColor(cell.ColorNumber(142)))); err != nil {
			return err
		}
		i++
		return nil
	})
	return t, nil
}

// newSparkLines creates two new sparklines displaying random values.
func newSparkLines(ctx context.Context) (*sparkline.SparkLine, *sparkline.SparkLine, error) {
	spGreen, err := sparkline.New(
		sparkline.Color(cell.ColorGreen),
	)
	if err != nil {
		return nil, nil, err
	}

	const max = 100
	go periodic(ctx, 250*time.Millisecond, func() error {
		v := int(rand.Int31n(max + 1))
		return spGreen.Add([]int{v})
	})

	spRed, err := sparkline.New(
		sparkline.Color(cell.ColorRed),
	)
	if err != nil {
		return nil, nil, err
	}
	go periodic(ctx, 500*time.Millisecond, func() error {
		v := int(rand.Int31n(max + 1))
		return spRed.Add([]int{v})
	})
	return spGreen, spRed, nil

}

// newGauge creates a demo Gauge widget.
func newGauge(ctx context.Context) (*gauge.Gauge, error) {
	g, err := gauge.New()
	if err != nil {
		return nil, err
	}

	const start = 35
	progress := start

	go periodic(ctx, 2*time.Second, func() error {
		if err := g.Percent(progress); err != nil {
			return err
		}
		progress++
		if progress > 100 {
			progress = start
		}
		return nil
	})
	return g, nil
}

// newDonut creates a demo Donut widget.
func newDonut(ctx context.Context) (*donut.Donut, error) {
	d, err := donut.New(donut.CellOpts(
		cell.FgColor(cell.ColorNumber(33))),
	)
	if err != nil {
		return nil, err
	}

	const start = 35
	progress := start

	go periodic(ctx, 500*time.Millisecond, func() error {
		if err := d.Percent(progress); err != nil {
			return err
		}
		progress++
		if progress > 100 {
			progress = start
		}
		return nil
	})
	return d, nil
}

// newBarChart returns a BarcChart that displays random values on multiple bars.
func newBarChart(ctx context.Context) (*barchart.BarChart, error) {
	bc, err := barchart.New(
		barchart.BarColors([]cell.Color{
			cell.ColorNumber(33),
			cell.ColorNumber(39),
			cell.ColorNumber(45),
			cell.ColorNumber(51),
			cell.ColorNumber(81),
			cell.ColorNumber(87),
		}),
		barchart.ValueColors([]cell.Color{
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
			cell.ColorBlack,
		}),
		barchart.ShowValues(),
	)
	if err != nil {
		return nil, err
	}

	const (
		bars = 6
		max  = 100
	)
	values := make([]int, bars)
	go periodic(ctx, 1*time.Second, func() error {
		for i := range values {
			values[i] = int(rand.Int31n(max + 1))
		}

		return bc.Values(values, max)
	})
	return bc, nil
}

// distance is a thread-safe int value used by the newSince method.
// Buttons write it and the line chart reads it.
type distance struct {
	v  int
	mu sync.Mutex
}

// add adds the provided value to the one stored.
func (d *distance) add(v int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.v += v
}

// get returns the current value.
func (d *distance) get() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.v
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
