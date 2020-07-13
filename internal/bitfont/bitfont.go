package bitfont

type Glyph struct {
	Mask []uint32
}

type Font struct {
	Width, Height int
	Glyphs        map[rune]Glyph
}
