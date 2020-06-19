package document

import (
	"bufio"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Document struct {
	Loop bool
	Sections []Section
}

type Section struct {
	Name string
	Duration time.Duration
}

func Read(file string) (raw string, doc Document) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		raw += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	err = yaml.Unmarshal([]byte(raw), &doc)
    if err != nil {
        panic(err)
	}
	
	return
}