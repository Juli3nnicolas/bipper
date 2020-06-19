package main

import "github.com/Juli3nnicolas/bipper/pkg/ui"

func main() {
	const bipFile string = "bip.mp3"
	const endBipFile string = "end_bip.mp3"
	const docFile string = "example.yaml"

	tui := ui.TermDashUI{}
	tui.Init(bipFile, endBipFile)
	tui.Run()
}