package text

import (
	"strings"
	"testing"

	"github.com/pbnjay/pixfont/cmd/fontgen/internal/parser"
)

func assertGlyphMatrix(t *testing.T, ch rune, m parser.Matrix, values ...uint32) {
	for i, v := range values {
		if m[i] != v {
			t.Errorf("character %c row %d: expected %032b got %032b", ch, i, v, m[i])
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
	tp := &textParser{}
	font, err := tp.Decode(strings.NewReader(document))
	if err != nil {
		t.Fatal(err)
	}

	if font.Width != 5 {
		t.Error("unexpected font width", font.Width)
	}

	if font.Height != 2 {
		t.Error("unexpected font height", font.Height)
	}

	assertGlyphMatrix(t, 'A', font.Glyphs['A'], 0b10101, 0b01010)
	assertGlyphMatrix(t, 'B', font.Glyphs['B'], 0b11100, 0b00011)
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
	tp := &textParser{}
	font, err := tp.Decode(strings.NewReader(document))
	if err != nil {
		t.Fatal(err)
	}

	if font.Width != 5 {
		t.Error("unexpected font width", font.Width)
	}

	if font.Height != 4 {
		t.Error("unexpected font height", font.Height)
	}

	assertGlyphMatrix(t, 'A', font.Glyphs['A'], 0b1, 0b0)
	assertGlyphMatrix(t, 'B', font.Glyphs['B'], 0b11100, 0b00011)
	assertGlyphMatrix(t, 'C', font.Glyphs['C'], 0b11100, 0b00011, 0b00011, 0b11100)
}
