package main

import (
	"fmt"

	"github.com/Juli3nnicolas/bipper/pkg/bipper"
)


func main() {
	const soundFile string = "bip.mp3"
	const docFile string = ""

	fmt.Println("Initialisaing bipper")
	bipper := bipper.Bipper{}
	bipper.Init(soundFile, docFile)

	fmt.Println("Starting bipper")
	bipper.Bip()
	bipper.Close()
	fmt.Println("Done.")
}

/*func main() {
	player := sound.NewPlayer()
	fmt.Println("loading file")
	player.Read("bip.mp3")
	fmt.Println("Playing file")
	player.Play()
	select{}
}*/