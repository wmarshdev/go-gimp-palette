package palette

import (
	"fmt"
	"image/color"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const validPaletteStrict = `GIMP Palette
Name: Valid Palette (Strict)
Columns: 2
#Comment line 1
#Comment line 2
0 127 255 color 1
0 0 0 color2

255 255 255
#Comment line 3
88 88 88 color 4
`

const validPaletteNoName = `GIMP Palette
Columns: 2
0 127 255
`

const validPaletteNoColumns = `GIMP Palette
Name: No Columns
0 127 255
`

const validPaletteMalformedColumns = `GIMP Palette
Name: Malformed Columns
Columns: foobarxxx
0 127 255
`

const validPaletteTruncatedRow = `GIMP Palette
Name: Valid Palette (truncated row)
Columns: 2
0 127 255
255 255
0 0 0
`

const validPaletteOutOfRange = `GIMP Palette
Name: Valid Palette (out of range)
Columns: 2
-100 999 255
`

const validPaletteMalformedRow = `GIMP Palette
Name: Valid Palette (malformed row)
Columns: 2
xx 127 255
255 255 255
0 0 0
`

type errReader struct {
	r io.Reader
}

func (r *errReader) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	if err == io.EOF {
		return 0, fmt.Errorf("read failed")
	}
	return
}

func TestReadPalette(t *testing.T) {
	t.Run("valid palette is parsed under strict validation", func(t *testing.T) {
		p, err := ReadPalette(strings.NewReader(validPaletteStrict), ParsingModeStrict)
		assert.NoError(t, err)
		if assert.NotNil(t, p) {
			assert.Equal(t, "Valid Palette (Strict)", p.Name)
			assert.Equal(t, 2, p.Columns)
			assert.Equal(t, []string{
				"Comment line 1",
				"Comment line 2",
				"Comment line 3",
			}, p.Comments)
			assert.Equal(t, []PaletteEntry{
				{"color 1", color.RGBA{0, 127, 255, 255}},
				{"color2", color.RGBA{0, 0, 0, 255}},
				{"", color.RGBA{255, 255, 255, 255}},
				{"color 4", color.RGBA{88, 88, 88, 255}},
			}, p.Entries)
		}
	})

	t.Run("valid palette is parsed under lenient validation", func(t *testing.T) {
		t.Run("no columns", func(t *testing.T) {
			p, err := ReadPalette(strings.NewReader(validPaletteNoColumns), ParsingModeLenient)
			assert.NoError(t, err)
			if assert.NotNil(t, p) {
				assert.Equal(t, "No Columns", p.Name)
				assert.Equal(t, 0, p.Columns)
				assert.Equal(t, []string{}, p.Comments)
				assert.Equal(t, []PaletteEntry{
					{"", color.RGBA{0, 127, 255, 255}},
				}, p.Entries)
			}
		})
		t.Run("no name", func(t *testing.T) {
			p, err := ReadPalette(strings.NewReader(validPaletteNoName), ParsingModeLenient)
			assert.NoError(t, err)
			if assert.NotNil(t, p) {
				assert.Equal(t, "", p.Name)
				assert.Equal(t, 2, p.Columns)
				assert.Equal(t, []string{}, p.Comments)
				assert.Equal(t, []PaletteEntry{
					{"", color.RGBA{0, 127, 255, 255}},
				}, p.Entries)
			}
		})
		t.Run("malformed columns", func(t *testing.T) {
			p, err := ReadPalette(strings.NewReader(validPaletteMalformedColumns), ParsingModeLenient)
			assert.NoError(t, err)
			if assert.NotNil(t, p) {
				assert.Equal(t, "Malformed Columns", p.Name)
				assert.Equal(t, 0, p.Columns)
				assert.Equal(t, []string{}, p.Comments)
				assert.Equal(t, []PaletteEntry{
					{"", color.RGBA{0, 127, 255, 255}},
				}, p.Entries)
			}
		})
		t.Run("truncated row (missing b value)", func(t *testing.T) {
			p, err := ReadPalette(strings.NewReader(validPaletteTruncatedRow), ParsingModeLenient)
			assert.NoError(t, err)
			if assert.NotNil(t, p) {
				assert.Equal(t, "Valid Palette (truncated row)", p.Name)
				assert.Equal(t, 2, p.Columns)
				assert.Equal(t, []string{}, p.Comments)
				assert.Equal(t, []PaletteEntry{
					{"", color.RGBA{0, 127, 255, 255}},
					{"", color.RGBA{255, 255, 0, 255}},
					{"", color.RGBA{0, 0, 0, 255}},
				}, p.Entries)
			}
		})
		t.Run("out of range row", func(t *testing.T) {
			p, err := ReadPalette(strings.NewReader(validPaletteOutOfRange), ParsingModeLenient)
			assert.NoError(t, err)
			if assert.NotNil(t, p) {
				assert.Equal(t, "Valid Palette (out of range)", p.Name)
				assert.Equal(t, 2, p.Columns)
				assert.Equal(t, []string{}, p.Comments)
				assert.Equal(t, []PaletteEntry{
					{"", color.RGBA{0, 255, 255, 255}},
				}, p.Entries)
			}
		})
		t.Run("malformed row", func(t *testing.T) {
			p, err := ReadPalette(strings.NewReader(validPaletteMalformedRow), ParsingModeLenient)
			assert.NoError(t, err)
			if assert.NotNil(t, p) {
				assert.Equal(t, "Valid Palette (malformed row)", p.Name)
				assert.Equal(t, 2, p.Columns)
				assert.Equal(t, []string{}, p.Comments)
				assert.Equal(t, []PaletteEntry{
					{"", color.RGBA{0, 127, 255, 255}},
					{"", color.RGBA{255, 255, 255, 255}},
					{"", color.RGBA{0, 0, 0, 255}},
				}, p.Entries)
			}
		})
	})

	t.Run("valid lenient palette is rejected under strict validation", func(t *testing.T) {
		t.Run("no columns", func(t *testing.T) {
			p, err := ReadPalette(strings.NewReader(validPaletteNoColumns), ParsingModeStrict)
			assert.ErrorIs(t, err, errMissingColumns)
			assert.Nil(t, p)
		})
		t.Run("no name", func(t *testing.T) {
			p, err := ReadPalette(strings.NewReader(validPaletteNoName), ParsingModeStrict)
			assert.ErrorIs(t, err, errMissingName)
			assert.Nil(t, p)
		})
		t.Run("malformed columns", func(t *testing.T) {
			p, err := ReadPalette(strings.NewReader(validPaletteMalformedColumns), ParsingModeStrict)
			assert.Error(t, err)
			assert.Nil(t, p)
		})
		t.Run("truncated row (missing b value)", func(t *testing.T) {
			p, err := ReadPalette(strings.NewReader(validPaletteTruncatedRow), ParsingModeStrict)
			assert.ErrorIs(t, err, errMissingField)
			assert.Nil(t, p)
		})
		t.Run("out of range row", func(t *testing.T) {
			p, err := ReadPalette(strings.NewReader(validPaletteOutOfRange), ParsingModeStrict)
			assert.ErrorIs(t, err, errOutOfRange)
			assert.Nil(t, p)
		})
		t.Run("malformed row", func(t *testing.T) {
			p, err := ReadPalette(strings.NewReader(validPaletteMalformedRow), ParsingModeStrict)
			assert.ErrorIs(t, err, errMalformedRow)
			assert.Nil(t, p)
		})
	})

	t.Run("bad header returns error (strict)", func(t *testing.T) {
		p, err := ReadPalette(strings.NewReader("foobar"), ParsingModeStrict)
		assert.ErrorIs(t, err, errBadHeader)
		assert.Nil(t, p)
	})

	t.Run("bad header returns error (lenient)", func(t *testing.T) {
		p, err := ReadPalette(strings.NewReader("foobar"), ParsingModeLenient)
		assert.ErrorIs(t, err, errBadHeader)
		assert.Nil(t, p)
	})

	t.Run("reader errors out", func(t *testing.T) {
		_, err := ReadPalette(&errReader{strings.NewReader(validPaletteStrict)}, ParsingModeLenient)
		assert.Error(t, err)
	})
}
