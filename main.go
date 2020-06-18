package main

import (
	"fmt"
	"time"

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
	fmt.Println("Done.")
}
