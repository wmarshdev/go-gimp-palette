package palette

import "image/color"

type Palette struct {
	Name     string
	Columns  int
	Comments []string
	Entries  []PaletteEntry
}

type PaletteEntry struct {
	Name  string
	Color color.Color
}
