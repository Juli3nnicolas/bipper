package document

import (
	"bufio"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Document struct {
	Loop     bool
	Sections []Section
	Dynamic
}

type Section struct {
	Name     string
	Duration time.Duration
}

// Dynamic is a struct containing values computed
// after the yaml doc has been read and the document
// struct hydrated
type Dynamic struct {
	// Total is the total time of every sections
	Total time.Duration
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

	setDynamics(&doc)

	return
}

// setDynamics sets all dynamics fields of Document doc
// Dynamic attributes are generated after the yaml document
// has been successfuly parsed
func setDynamics(doc *Document) {
	for _, s := range doc.Sections {
		doc.Total += s.Duration
	}
}
