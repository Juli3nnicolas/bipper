package bipper

import (
	"fmt"
	"time"

	"github.com/Juli3nnicolas/bipper/pkg/document"
	"github.com/Juli3nnicolas/bipper/pkg/sound"
)

type BipperOutput struct {
	Msg 		chan string
	SectionName chan string
	Remaining 	chan time.Duration
}

type Bipper struct {
	Output		BipperOutput
	player 		sound.Player
	endPlayer 	sound.Player
	doc 		document.Document
}

func (o *Bipper) Init(bipFile, endBipFile, docFile string) {
	o.Output.Msg = make(chan string)
	o.Output.SectionName = make(chan string)
	o.Output.Remaining = make(chan time.Duration)
	
	o.player = sound.NewPlayer()
	o.player.Read(bipFile)
	
	o.endPlayer = sound.NewPlayer()
	o.endPlayer.Read(endBipFile)

	o.doc = document.Read(docFile)
}

func (o *Bipper) Bip() {
	loop := true
	tick := time.Tick(time.Second)

	for loop {
		for _, section := range o.doc.Sections {
			o.Output.Msg <- fmt.Sprintf("\nRunning section %s lasting %v\n", section.Name, section.Duration)
			o.Output.SectionName <- section.Name
			
			var timer time.Time
			alarm := time.After(section.Duration)

			countingDown := true
			for countingDown {
				select {
					case <-tick:
						timer = timer.Add(time.Second)
						duration := time.Time{}.Add(section.Duration)
						remaining := duration.Sub(timer)
						remainingSec := remaining.Seconds()

						o.Output.Remaining <- remaining
						if remainingSec >= 1.0 && remainingSec <= 3.0 {
							o.player.Play()
							o.Output.Msg <- fmt.Sprintf("%s: %.0f\n", section.Name, remainingSec)
						}

					case <-alarm:
						o.endPlayer.Play()
						o.Output.Msg <- fmt.Sprintf("Section %s is over\n", section.Name)
						countingDown = false
				}
			}
		}
		loop = o.doc.Loop
	}
}

func (o *Bipper) Close() {
	o.player.Close()
}