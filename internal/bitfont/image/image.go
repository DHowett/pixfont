package image

import (
	"image"
	"image/color"
	"io"
	"unicode/utf8"

	"github.com/pbnjay/pixfont/internal/bitfont"
)

type Options struct {
	Offset image.Point
	Size   image.Point
}

func Decode(r io.Reader, alphabet string, options *Options) (*bitfont.Font, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	var offset image.Point
	var size image.Point
	if options != nil {
		offset = options.Offset
		size = options.Size
	}

	bounds := img.Bounds()
	bounds.Min = offset
	if size.X != 0 {
		bounds.Max.X = bounds.Min.X + size.X
	}

	if size.Y != 0 {
		bounds.Max.Y = bounds.Min.Y + size.Y
	}

	glyphs := make(map[rune]bitfont.Glyph)
	maxWidth := 0

	// generate a greyscale histogram of the image
	pxc := 0
	clrs := make(map[uint8]int)
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			c := img.At(x, y)
			gc := color.GrayModel.Convert(c).(color.Gray)
			clrs[gc.Y]++
			pxc++
		}
	}

	// find a threshold pixel count for what colors to ignore as background
	// (ie assumes background image is fairly solid and colors occur much
	//  more often than font colors)
	pxt := pxc
	pxd := 0
	for pxd < (pxc/2) && pxt > 0 {
		pxt /= 2
		pxd = 0
		for _, n := range clrs {
			if n > pxt {
				pxd += n
			}
		}
	}

	// scan across the image in the crop region, saving pixels as you go.
	// if at any point we see an "empty" column of pixels, we assume it
	// is a character boundary and move to the next alphabet letter.
	curAlpha := alphabet
	curWidth := 0
	curMatrix := make([]uint32, 0, 16)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		curWidth++
		isEmpty := true
		ay := 0
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			c := img.At(x, y)
			gc := color.GrayModel.Convert(c).(color.Gray)
			if len(curMatrix) < (ay + 1) {
				curMatrix = append(curMatrix, uint32(0))
			}
			curMatrix[ay] >>= 1
			if clrs[gc.Y] <= pxt {
				curMatrix[ay] |= 0x80000000
				isEmpty = false
			}
			ay++
		}

		if isEmpty {
			if len(curMatrix) != 0 {
				if len(curAlpha) > 0 {
					curWidth-- // remove last blank column
					for yy, _ := range curMatrix {
						curMatrix[yy] >>= 32 - curWidth - 1
					}
					r, nbytes := utf8.DecodeRuneInString(curAlpha)
					glyphs[r] = bitfont.Glyph{Mask: curMatrix}
					curAlpha = curAlpha[nbytes:]
				}
				if curWidth > maxWidth {
					maxWidth = curWidth
				}
			}
			curWidth = 0
			curMatrix = make([]uint32, 0, 16)
		}
	}

	// Stuff the last glyph (if it exists) into the map as well
	if len(curMatrix) != 0 && len(curAlpha) > 0 {
		// we're not trimming any more columns, as we only got here because
		// there was *no* isEmpty column that would have made us emit the last
		// glyph
		// still, we must scoot the glyph to the LSB
		for yy, _ := range curMatrix {
			curMatrix[yy] >>= 32 - curWidth
		}
		r, _ := utf8.DecodeRuneInString(curAlpha)
		glyphs[r] = bitfont.Glyph{Mask: curMatrix}
		if curWidth > maxWidth {
			maxWidth = curWidth
		}
	}

	return &bitfont.Font{
		Width:  maxWidth,
		Height: bounds.Dy(),
		Glyphs: glyphs,
	}, nil
}
