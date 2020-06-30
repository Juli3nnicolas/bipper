package bipper

import (
	"fmt"
	"time"

	"github.com/Juli3nnicolas/bipper/pkg/document"
	"github.com/Juli3nnicolas/bipper/pkg/sound"
)

type BipperOutput struct {
	Msg            chan string
	Section        chan document.Section
	RawDoc         chan string
	Remaining      chan time.Duration
	TotalRemaining chan time.Duration
}

type BipperInput struct {
	TogglePause chan bool
}

type Bipper struct {
	Input     BipperInput
	Output    BipperOutput
	player    sound.Player
	endPlayer sound.Player
	rawDoc    string
	doc       document.Document
}

func (o *Bipper) Init(bipFile, endBipFile, docFile string) (err error) {
	o.Input.TogglePause = make(chan bool)

	o.Output.Msg = make(chan string)
	o.Output.Section = make(chan document.Section)
	o.Output.RawDoc = make(chan string)
	o.Output.Remaining = make(chan time.Duration)
	o.Output.TotalRemaining = make(chan time.Duration)

	o.player = sound.NewPlayer()
	o.player.Read(bipFile)

	o.endPlayer = sound.NewPlayer()
	o.endPlayer.Read(endBipFile)

	o.rawDoc, o.doc, err = document.Read(docFile)
	return
}

func (o *Bipper) Bip() {
	o.Output.RawDoc <- o.rawDoc

	loop := true
	tick := time.Tick(time.Second)
	totalRemaining := o.doc.Total
	pause := false

	for loop {
		for _, section := range o.doc.Sections {
			o.Output.Msg <- fmt.Sprintf("\nRunning section %s lasting %v\n", section.Name, section.Duration)
			o.Output.Section <- section

			var timer time.Time

			countingDown := true
			for countingDown {
				select {
				case <-o.Input.TogglePause:
					pause = !pause

				case <-tick:
					if !pause {
						timer = timer.Add(time.Second)
						duration := time.Time{}.Add(section.Duration)
						remaining := duration.Sub(timer)
						remainingSec := remaining.Seconds()
						totalRemaining -= time.Second

						// When the time is over - play end bip and resume section processing (exit select)
						if remainingSec <= 0 {
							o.Output.Remaining <- 0
							o.Output.TotalRemaining <- totalRemaining
							o.endPlayer.Play()
							o.Output.Msg <- fmt.Sprintf("Section %s is over\n", section.Name)
							countingDown = false
							break
						}

						o.Output.Remaining <- remaining
						if remainingSec >= 1.0 && remainingSec <= 3.0 {
							o.player.Play()
							o.Output.Msg <- fmt.Sprintf("%s: %.0f\n", section.Name, remainingSec)
						}

						o.Output.TotalRemaining <- totalRemaining
					}
				}
			}
		}
		loop = o.doc.Loop
	}
}

func (o *Bipper) Close() {
	if o.player != nil {
		o.player.Close()
	}
}
