package text

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"regexp"
	"unicode/utf8"

	"github.com/pbnjay/pixfont/cmd/fontgen/internal/parser"
)

type textParser struct {
}

var _ parser.Parser = &textParser{}

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

func (*textParser) Decode(r io.Reader) (*parser.Font, error) {
	input, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`[^\n]*\n`)
	hh, maxHeight, maxWidth := 0, 0, 0
	lastCh := rune(0)

	glyphs := make(map[rune]parser.Matrix)

	for _, bline := range re.FindAll(input, -1) {
		line := string(bline)
		c, pixoffs := utf8.DecodeRuneInString(line)
		pixoffs += 3
		if lastCh != c {
			hh = len(glyphs[lastCh])
			if hh > maxHeight {
				maxHeight = hh
			}
		}
		ww := len(line) - (pixoffs + 2)
		if ww > maxWidth {
			maxWidth = ww
		}
		glyphs[c] = append(glyphs[c], textRepresentationToBits(line[pixoffs:pixoffs+ww]))
		lastCh = c
	}

	hh = len(glyphs[lastCh])
	if hh > maxHeight {
		// this may be left over from the last character
		maxHeight = hh
	}

	return &parser.Font{
		Width:  maxWidth,
		Height: maxHeight,
		Glyphs: glyphs,
	}, nil
}

func NewParser() parser.Parser {
	return &textParser{}
}
