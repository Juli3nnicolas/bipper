package main

import (
	//"fmt"
	"time"

	//"github.com/Juli3nnicolas/bipper/pkg/bipper"
	"github.com/Juli3nnicolas/bipper/pkg/ui"
)


func main() {
	const bipFile string = "bip.mp3"
	const endBipFile string = "end_bip.mp3"
	const docFile string = "example.yaml"

	//fmt.Println("Initialisaing bipper")
	//bipper := bipper.Bipper{}
	//bipper.Init(bipFile, endBipFile, docFile)

	/*go func () {
		//fmt.Println("Starting bipper")
		bipper.Bip()
		bipper.Close()
		//fmt.Println("\nDone. Press ctrl + C to exit.")
	}()*/

	tui := ui.TermDashUI{}
	tui.Init(bipFile, endBipFile)
	tui.Run()
	
	// Leave the app open to play all remaining sounds
	for {
		time.Sleep(3*time.Second)
	}
}