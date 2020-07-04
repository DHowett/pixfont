package image

import (
	"image"
	"os"
	"path/filepath"
	"testing"
	_ "image/png"

	"github.com/pbnjay/pixfont/cmd/fontgen/internal/parser"
)

func assertGlyphMatrix(t *testing.T, ch rune, m parser.Matrix, values ...uint32) {
	for i, v := range values {
		if m[i] != v {
			t.Errorf("character %c row %d: expected %032b got %032b", ch, i, v, m[i])
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

	p := &imageParser{
		alphabet: "ABC",
	}
	font, err := p.Decode(f)
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

	assertGlyphMatrix(t, 'A', font.Glyphs['A'], 0b010, 0b101, 0b111, 0b101, 0b101)
	assertGlyphMatrix(t, 'B', font.Glyphs['B'], 0b011, 0b101, 0b011, 0b101, 0b011)
	assertGlyphMatrix(t, 'C', font.Glyphs['C'], 0b0110, 0b1001, 0b0001, 0b1001, 0b0110)
}

func TestParseOffset(t *testing.T) {
	f := testOpenFile(t, "abc.png")
	defer f.Close()

	p := &imageParser{
		alphabet: "BC",
		// I only want BC
		offset: image.Point{4, 0},
	}
	font, err := p.Decode(f)
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

	assertGlyphMatrix(t, 'B', font.Glyphs['B'], 0b011, 0b101, 0b011, 0b101, 0b011)
	assertGlyphMatrix(t, 'C', font.Glyphs['C'], 0b0110, 0b1001, 0b0001, 0b1001, 0b0110)
}

func TestParseWindow(t *testing.T) {
	f := testOpenFile(t, "abc.png")
	defer f.Close()

	p := &imageParser{
		alphabet: "BC",
		// I only want B
		offset: image.Point{4, 0},
		size:   image.Point{3, 0}, // y=0 == y=max
	}
	font, err := p.Decode(f)
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

	assertGlyphMatrix(t, 'B', font.Glyphs['B'], 0b011, 0b101, 0b011, 0b101, 0b011)
}
