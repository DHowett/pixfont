package text

import (
	"bufio"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/pbnjay/pixfont/internal/bitfont"
)

// Transforms a 0-32-character-long string of spaces and Xs into a uint32
func textRepresentationToBits(t string) uint32 {
	var o uint32
	for i := 0; i < len(t); i++ {
		o >>= 1
		if t[i] == 'X' {
			o |= 0x80000000
		}
	}
	o >>= 32 - len(t)
	return o
}

func Decode(r io.Reader) (*bitfont.Font, error) {
	maxHeight, maxWidth := 0, 0
	lastCh := rune(0)

	glyphs := make(map[rune]bitfont.Glyph)

	scanner := bufio.NewScanner(r)
	var g bitfont.Glyph
	for scanner.Scan() {
		line := scanner.Text()
		c, pixoffs := utf8.DecodeRuneInString(line)
		pixoffs += 3
		ww := strings.IndexRune(line[pixoffs:], ']')

		if c != lastCh {
			if lastCh != 0 {
				if hh := len(g.Mask); hh > maxHeight {
					maxHeight = hh
				}
				glyphs[lastCh] = g
				g = bitfont.Glyph{}
			}
		}

		if ww > maxWidth {
			maxWidth = ww
		}
		g.Mask = append(g.Mask, textRepresentationToBits(line[pixoffs:pixoffs+ww]))
		lastCh = c
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	glyphs[lastCh] = g

	if hh := len(glyphs[lastCh].Mask); hh > maxHeight {
		// this may be left over from the last character
		maxHeight = hh
	}

	return &bitfont.Font{
		Width:  maxWidth,
		Height: maxHeight,
		Glyphs: glyphs,
	}, nil
}
