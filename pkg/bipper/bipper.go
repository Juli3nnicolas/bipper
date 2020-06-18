package bipper

import (
	"fmt"
	"time"

	"github.com/Juli3nnicolas/bipper/pkg/sound"
)

type Document struct {
	Loop bool
	Sections []Section
}

type Section struct {
	Name string
	Duration time.Duration
}

type Bipper struct {
	player sound.Player
	endPlayer sound.Player
	doc Document
}

func (o *Bipper) Init(bipFile, endBipFile, docFile string) {
	o.player = sound.NewPlayer()
	o.player.Read(bipFile)
	
	o.endPlayer = sound.NewPlayer()
	o.endPlayer.Read(endBipFile)

	o.doc = Document{
		Loop: false,
		Sections: []Section{
			{"first section", 10*time.Second},
			{"second section", 10*time.Second},
		},
	}
}

func (o *Bipper) Bip() {
	loop := true
	tick := time.Tick(time.Second)

	for loop {
		for _, section := range o.doc.Sections {
			fmt.Printf("\nRunning section %s lasting %v\n", section.Name, section.Duration)
			var timer time.Time
			alarm := time.After(section.Duration)

			countingDown := true
			for countingDown {
				select {
					case <-tick:
						timer = timer.Add(time.Second)
						duration := time.Time{}.Add(section.Duration)
						remaining := duration.Sub(timer).Seconds()

						if remaining >= 1.0 && remaining <= 3.0 {
							o.player.Play()
							fmt.Printf("%s: %.0f\n", section.Name, remaining)
						}

					case <-alarm:
						o.endPlayer.Play()
						fmt.Printf("Section %s is over\n", section.Name)
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