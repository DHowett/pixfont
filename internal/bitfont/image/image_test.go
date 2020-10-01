package image

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"
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

func testOpenFile(t *testing.T, name string) *os.File {
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("error opening %s: %v", name, err)
	}
	return f
}

func TestParseNoOffset(t *testing.T) {
	f := testOpenFile(t, "abc.png")
	defer f.Close()

	font, err := Decode(f, "ABC", nil)
	if err != nil {
		t.Fatal(err)
	}

	if font.Width != 4 {
		t.Error("unexpected font width", font.Width)
	}

	if font.Height != 5 {
		t.Error("unexpected font height", font.Height)
	}

	if len(font.Glyphs) != 3 {
		t.Error("unexpected glyph count", len(font.Glyphs))
	}

	assertGlyphMask(t, 'A', font.Glyphs['A'], 0b010, 0b101, 0b111, 0b101, 0b101)
	assertGlyphMask(t, 'B', font.Glyphs['B'], 0b011, 0b101, 0b011, 0b101, 0b011)
	assertGlyphMask(t, 'C', font.Glyphs['C'], 0b0110, 0b1001, 0b0001, 0b1001, 0b0110)
}

func TestParseOffset(t *testing.T) {
	f := testOpenFile(t, "abc.png")
	defer f.Close()

	font, err := Decode(f, "BC", &Options{
		// I only want BC so crop out A
		Offset: image.Point{4, 0},
	})
	if err != nil {
		t.Fatal(err)
	}

	if font.Width != 4 {
		t.Error("unexpected font width", font.Width)
	}

	if font.Height != 5 {
		t.Error("unexpected font height", font.Height)
	}

	if len(font.Glyphs) != 2 {
		t.Error("unexpected glyph count", len(font.Glyphs))
	}

	assertGlyphMask(t, 'B', font.Glyphs['B'], 0b011, 0b101, 0b011, 0b101, 0b011)
	assertGlyphMask(t, 'C', font.Glyphs['C'], 0b0110, 0b1001, 0b0001, 0b1001, 0b0110)
}

func TestParseWindow(t *testing.T) {
	f := testOpenFile(t, "abc.png")
	defer f.Close()

	font, err := Decode(f, "BC", &Options{
		// we only want B, so crop out A and C
		Offset: image.Point{4, 0},
		Size:   image.Point{3, 0},
	})
	if err != nil {
		t.Fatal(err)
	}

	if font.Width != 3 {
		t.Error("unexpected font width", font.Width)
	}

	if font.Height != 5 {
		t.Error("unexpected font height", font.Height)
	}

	if len(font.Glyphs) != 1 {
		t.Error("unexpected glyph count", len(font.Glyphs))
	}

	assertGlyphMask(t, 'B', font.Glyphs['B'], 0b011, 0b101, 0b011, 0b101, 0b011)
}
