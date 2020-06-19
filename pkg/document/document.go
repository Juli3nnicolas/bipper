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

func Read(file string) (raw string, doc Document, err error) {
	f, err := os.Open(file)
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		raw += scanner.Text() + "\n"
	}

	if err = scanner.Err(); err != nil {
		return
	}

	err = yaml.Unmarshal([]byte(raw), &doc)
    if err != nil {
        return
	}
	
	return
}