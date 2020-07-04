// fontgen is a commandline tool for generating pixel fonts supported by pixfont.
// First is to create an image of your pixel font in your favorite graphics
// program with your set of supported characters. Ex:
//
//      ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789
//
// Ensure that there is a solid color background, single-color font pixels (i.e.
// no anti-aliasing), and that a column of pixels separate each character to
// ensure best results. Then simply run:
//
//      ./fontgen -img mypixelfont.png -o myfont
//
// Add myfont.go to your project, then just use Font.DrawString(...) to add
// text to your image!
//
package main

import (
	"flag"
	"fmt"
	"go/format"
	"image"
	"os"
	"sort"

	"github.com/pbnjay/pixfont/cmd/fontgen/internal/parser"
	pimg "github.com/pbnjay/pixfont/cmd/fontgen/internal/parser/image"
	ptext "github.com/pbnjay/pixfont/cmd/fontgen/internal/parser/text"

	// used by the image parser
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/pbnjay/pixfont"
)

var (
	imageName = flag.String("img", "", "image file to extract pixel font from")
	startY    = flag.Int("y", 0, "starting Y position")
	height    = flag.Int("h", 0, "chop height")
	startX    = flag.Int("x", 0, "starting X position")
	width     = flag.Int("w", 0, "chop width")
	alphabet  = flag.String("a", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789", "alphabet to extract")
	varWidth  = flag.Bool("v", false, "produce variable width font")

	textName = flag.String("txt", "", "text file to extract pixel font from")
	outName  = flag.String("o", "", "package name to create (becomes <myfont>.go)")
)

// packFont packs a font (where each glyph is stored in its own uint32 matrix)
// and returns a packed binary representation (with multiple glyphs per uint32)
// plus a location mapping
func packFont(font *parser.Font) ([]uint32, map[rune]uint16) {
	cm := make(map[rune]uint16)

	chs := make([]int, 0, len(font.Glyphs))
	for ch, _ := range font.Glyphs {
		chs = append(chs, int(ch))
	}
	sort.IntSlice(chs).Sort()

	w, h := font.Width, font.Height

	// convert from simple character encoding to packed bitfield
	// NB fonts should be at most 32 pixels wide to fit in the uint32
	//    (height is limited to uint8 255)
	//
	// This packed representation stores 1-4 glyphs in a single uint32 (per line).
	// For efficiency, each glyph must be 8-bit aligned. Glyphs are stored "backwards"
	// (leftmost pixel in LSB).
	// Glyphs that will not fit in their entirety will be pushed to the next uint32.
	//
	// For example:
	// An 8-pixel font can store 4 glyphs using one uint32 per line.
	// A 9-pixel font can only store 2, because 9-bit values must be
	// byte-aligned.
	// A 17-pixel font can only store 1, because it is impossible to
	// align two 17-bit values (totalling 34 bits) in 32.
	//
	// Lines are stored in consecutive uint32s.
	//
	//         24      16       8       0
	//          |       |       |       |
	// 0     DDDD    CCC     BBBB     A   == 0b00001111000011100000111100000100 == 0x0f0e0f04
	// 1    D   D   C   C   B   B    A A  == 0b00010001000100010001000100001010 == 0x1111110a
	// 2    D   D       C    BBBB   A   A == 0b00010001000000010000111100010001 == 0x11010f11
	// 3    D   D   C   C   B   B   AAAAA == 0b00010001000100010001000100011111 == 0x1111111f
	// 4     DDDD    CCC     BBBB   A   A == 0b00001111000011100000111100010001 == 0x0f0e0f11
	// 5                            EEEEE == 0b00000000000000000000000000011111 == 0x0000001f
	// 6                                E == 0b00000000000000000000000000000001 == 0x00000001
	// 7                             EEEE == 0b00000000000000000000000000001111 == 0x0000000f
	// 8                                E == 0b00000000000000000000000000000001 == 0x00000001
	// 9                            EEEEE == 0b00000000000000000000000000011111 == 0x0000001f

	u8PerCh := ((w - 1) >> 3) + 1 // 0-8 take up 1 byte, 9-16 take up 2, 17-24 take up 3, 24+ take up 4
	chPerU32 := 4 / u8PerCh       // we can fit 4, 2 or 1 glyphs per u32
	spacing := 4 / chPerU32       // we must skip 1, 2, or 4 8-bit units between each glyph start

	costPerLine := (len(chs) + chPerU32 - 1) / chPerU32 // #of whole u32 per horizontal line in font
	costTotal := h * costPerLine                        // #of whole u32s required for the whole font

	encoded := make([]uint32, costTotal)

	// i8 tracks the number of 8-bit units we've skipped
	var i8 int
	for _, c := range chs {
		matrix := font.Glyphs[rune(c)]

		i32 := (i8 >> 2) * h // i32 is the index into encoded for the u32 for this char
		dist := i8 & 0b11    // how many u8 units into the u32 we're offset
		cm[rune(c)] = uint16((i32 << 2) | dist)

		for y, line := range matrix {
			encoded[i32+y] |= (line << (8 * dist))
		}

		i8 += spacing
	}
	return encoded, cm
}

func generatePixFont(name string, v bool, font *parser.Font) {
	template := `
		package %s

		import "github.com/pbnjay/pixfont"

		var Font *pixfont.PixFont

		func init() {
			charMap := %#v
			data := %#v
			Font = pixfont.NewPixFont(%d, %d, charMap, data)
			Font.SetVariableWidth(%t)
		}
	`

	encoded, cm := packFont(font)

	fnt := pixfont.NewPixFont(uint8(font.Width), uint8(font.Height), cm, encoded)
	fnt.SetVariableWidth(v)

	f, err := os.OpenFile(name+".go", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	// draw a comment header using the new font
	sd := &pixfont.StringDrawable{}
	fnt.DrawString(sd, 0, 0, name, nil)
	fmt.Fprintln(f, sd.PrefixString("// "))

	// create the code from the template and go fmt it
	code := fmt.Sprintf(template, name, cm, encoded, font.Width, font.Height, v)
	bcode, _ := format.Source([]byte(code))
	fmt.Fprintln(f, string(bcode))

	f.Close()
}

func bitsToString(b uint32, w int) string {
	s := ""
	for i := 0; i < w; i++ {
		if (b & 1) == 1 {
			s += "X"
		} else {
			s += " "
		}
		b >>= 1
	}
	return s
}

func dumpFont(font *parser.Font) {
	chs := make([]int, 0, len(font.Glyphs))
	for ch, _ := range font.Glyphs {
		chs = append(chs, int(ch))
	}
	sort.IntSlice(chs).Sort()

	for _, ch := range chs {
		for _, line := range font.Glyphs[rune(ch)] {
			fmt.Printf("%c  [%s]\n", rune(ch), bitsToString(line, font.Width))
		}
	}
}

func main() {
	flag.Parse()

	var filename string
	var p parser.Parser
	if *imageName != "" {
		filename = *imageName
		offset := image.Point{}
		size := image.Point{}
		if startX != nil {
			offset.X = *startX
		}
		if startY != nil {
			offset.Y = *startY
		}
		if width != nil {
			size.X = *width
		}
		if height != nil {
			size.Y = *height
		}
		imageParser := pimg.NewParser(*alphabet, offset, size)
		p = imageParser
	} else if *textName != "" {
		filename = *textName
		p = ptext.NewParser()
	}

	if p == nil {
		fmt.Fprintln(os.Stderr, "-img or -txt should be provided")
		flag.Usage()
		return
	}

	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to open file:", err)
		os.Exit(1)
	}
	defer f.Close()

	font, err := p.Decode(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error parsing file:", err)
		os.Exit(1)
	}

	if *outName != "" {
		generatePixFont(*outName, *varWidth, font)
		fmt.Fprintln(os.Stderr, "Created package file:", *outName+".go")
	} else {
		// dump a text representation of the font to stdout
		dumpFont(font)
	}
}
