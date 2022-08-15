package palette

import (
	"errors"
	"fmt"
	"image/color"
	"io"
	"math"
	"strings"
	"unicode"

	"go.uber.org/multierr"
)

var errBadHeader = errors.New("missing magic header")
var errMissingName = errors.New("missing palette name")
var errMissingColumns = errors.New("missing palette columns")
var errMissingField = errors.New("missing field")
var errOutOfRange = errors.New("value out of range")
var errMalformedRow = errors.New("row is malformed")

const magicHeader = "GIMP Palette"

type ParsingMode int

const (
	ParsingModeLenient ParsingMode = iota
	ParsingModeStrict
)

func ReadPalette(r io.Reader, parsingMode ParsingMode) (p *Palette, err error) {
	done := make(chan struct{})

	// read lines
	linesCh, linesErrCh := readLines(done, r)

	// combine errors
	defer multierr.AppendInvoke(&err, multierr.Invoke(func() (err error) {
		for linesErr := range linesErrCh {
			multierr.AppendInto(&err, linesErr)
		}
		return
	}),
	)

	defer close(done)

	// expect magic header
	if line, ok := <-linesCh; !ok || line != magicHeader {
		multierr.AppendInto(&err, errBadHeader)
		return
	}

	// expect name
	var name string
	if line, ok := <-linesCh; !ok || !strings.HasPrefix(line, "Name: ") {
		if parsingMode == ParsingModeStrict {
			multierr.AppendInto(&err, errMissingName)
			return
		}
		linesCh = putBack(line, linesCh)
	} else {
		name = strings.TrimSpace(line[len("Name:"):])
	}

	// expect columns
	var columns int
	if line, ok := <-linesCh; !ok || !strings.HasPrefix(line, "Columns: ") {
		if parsingMode == ParsingModeStrict {
			multierr.AppendInto(&err, errMissingColumns)
			return
		}
		linesCh = putBack(line, linesCh)
	} else {
		if _, scanErr := fmt.Sscanf(strings.TrimSpace(line[len("Columns: "):]), "%d", &columns); scanErr != nil {
			if parsingMode == ParsingModeStrict {
				multierr.AppendInto(&err, fmt.Errorf("bad columns entry: %w", scanErr))
				return
			}
		}
	}

	// process remaining lines
	comments := []string{}
	entries := []PaletteEntry{}

	for line := range linesCh {
		line := strings.TrimSpace(line)
		if line == "" {
			// blank
		} else if strings.HasPrefix(line, "#") {
			// comment
			comments = append(comments, line[1:])
		} else {
			// entry
			if entry, parseErr := parseRow(line, parsingMode); parseErr != nil {
				multierr.AppendInto(&err, parseErr)
				return
			} else {
				entries = append(entries, *entry)
			}
		}
	}

	p = &Palette{name, columns, comments, entries}

	return
}

func parseRow(line string, parsingMode ParsingMode) (entry *PaletteEntry, err error) {
	color := color.RGBA{0, 0, 0, 255}
	var entryName string
	splitCount := 0
	fields := strings.FieldsFunc(line, func(r rune) bool {
		if unicode.IsSpace(r) && splitCount < 3 {
			splitCount++
			return true
		}
		return false
	})

	if parsingMode == ParsingModeStrict && len(fields) < 3 {
		multierr.AppendInto(&err, errMissingField)
		return
	}

	var fieldErr error
	processField := func(field string) uint8 {
		var v int
		if _, err := fmt.Sscanf(field, "%d", &v); err != nil {
			if parsingMode == ParsingModeStrict {
				multierr.AppendInto(&fieldErr, errMalformedRow)
			}
		}
		if v < 0 || v > 255 {
			if parsingMode == ParsingModeStrict {
				multierr.AppendInto(&fieldErr, errOutOfRange)
			}
			v = int(math.Trunc(math.Max(0., math.Min(255., float64(v)))))
		}

		return uint8(v & 0xFF)
	}
	// r
	if len(fields) >= 1 {
		color.R = processField(fields[0])
	}
	// g
	if len(fields) >= 2 {
		color.G = processField(fields[1])
	}
	// b
	if len(fields) >= 3 {
		color.B = processField(fields[2])
	}

	if multierr.AppendInto(&err, fieldErr) {
		return
	}

	if len(fields) >= 4 {
		entryName = fields[3]
	}

	entry = &PaletteEntry{entryName, color}

	return
}
