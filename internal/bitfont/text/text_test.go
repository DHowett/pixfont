package text

import (
	"strings"
	"testing"

	"github.com/pbnjay/pixfont/internal/bitfont"
)

func assertGlyphMask(t *testing.T, ch rune, g bitfont.Glyph, values ...uint32) {
	for i, v := range values {
		if g.Mask[i] != v {
			t.Errorf("character %c row %d: expected %032b got %032b", ch, i, v, g.Mask[i])
		}
	}
}

func TestParseFixed(t *testing.T) {
	// These glyphs are uniformly 5x2.
	var document = `A  [X X X]
A  [ X X ]
B  [  XXX]
B  [XX   ]
`
	font, err := Decode(strings.NewReader(document))
	if err != nil {
		t.Fatal(err)
	}

	if font.Width != 5 {
		t.Error("unexpected font width", font.Width)
	}

	if font.Height != 2 {
		t.Error("unexpected font height", font.Height)
	}

	assertGlyphMask(t, 'A', font.Glyphs['A'], 0b10101, 0b01010)
	assertGlyphMask(t, 'B', font.Glyphs['B'], 0b11100, 0b00011)
}

func TestParseVariable(t *testing.T) {
	// These glyphs vary in both width and height.
	var document = `A  [X]
A  [ ]
B  [  XXX]
B  [XX   ]
C  [  XXX]
C  [XX   ]
C  [XX   ]
C  [  XXX]
`
	font, err := Decode(strings.NewReader(document))
	if err != nil {
		t.Fatal(err)
	}

	if font.Width != 5 {
		t.Error("unexpected font width", font.Width)
	}

	if font.Height != 4 {
		t.Error("unexpected font height", font.Height)
	}

	assertGlyphMask(t, 'A', font.Glyphs['A'], 0b1, 0b0)
	assertGlyphMask(t, 'B', font.Glyphs['B'], 0b11100, 0b00011)
	assertGlyphMask(t, 'C', font.Glyphs['C'], 0b11100, 0b00011, 0b00011, 0b11100)
}
