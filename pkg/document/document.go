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

func Read(file string) Document {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(f)
	raw := ""

	for scanner.Scan() {
		raw += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	doc := Document{}
	err = yaml.Unmarshal([]byte(raw), &doc)
    if err != nil {
        panic(err)
	}
	
	return doc
}