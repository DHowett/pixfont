package bdf

import (
	"io"
	"math/bits"

	"github.com/pbnjay/pixfont/cmd/fontgen/internal/parser"
)

type bdfParser struct{}

func (p *bdfParser) Decode(r io.Reader) (*parser.Font, error) {
	bf, err := OpenBDF(r)
	if err != nil {
		return nil, err
	}

	// OpenBDF has already handled the font bounding box
	// and all that matters is the glyph's bbox
	glyphs := make(map[rune]parser.Matrix, len(bf.Glyphs))
	width, height := -1, -1
	for _, bg := range bf.Glyphs {
		xoff, yoff := bg.BoundingBox[2], bg.BoundingBox[3]
		h := len(bg.Bitmap) + yoff
		matrix := make(parser.Matrix, h)
		for i, b := range bg.Bitmap {
			// bitmap occupies the top WIDTH bits of the bottom nearest increment of 8
			o := 32 - ((((bg.BoundingBox[0] - 1) / 8) + 1) * 8)
			// move to the top of the u32 (left pixel = MSB)
			line := b << o
			// flip it (left pixel = LSB)
			line = bits.Reverse32(line)
			// include the x offset (move bitmap left, bits right)
			line <<= xoff
			matrix[yoff+i] = line
		}
		hh := len(bg.Bitmap) + yoff
		if width < bg.Width {
			width = bg.Width
		}
		if height < hh {
			height = hh
		}
		glyphs[bg.Encoding] = matrix
	}

	return &parser.Font{
		Width:  width,
		Height: height,
		Glyphs: glyphs,
	}, nil
}

func NewParser() parser.Parser {
	return &bdfParser{}
}
