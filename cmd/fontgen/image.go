package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
	"unicode/utf8"
)

func processImage(filename string) (allLetters map[rune]map[int]string, maxWidth int) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return nil, 0
	}
	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		return nil, 0
	}
	if *width == 0 {
		*width = img.Bounds().Dx() - *startX
	}
	if *height == 0 {
		*height = img.Bounds().Dy() - *startY
	}
	allLetters = make(map[rune]map[int]string)
	maxWidth = 0

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
	curAlpha := *alphabet
	curWidth := 0
	curLetter := make(map[int]string)
	for x := *startX; x < *startX+*width; x++ {
		curWidth++
		isEmpty := true
		ay := 0
		for y := *startY; y < *startY+*height; y++ {
			c := img.At(x, y)
			gc := color.GrayModel.Convert(c).(color.Gray)
			if clrs[gc.Y] <= pxt {
				if _, haveDots := curLetter[ay]; !haveDots {
					curLetter[ay] = strings.Repeat(" ", curWidth-1)
				}
				curLetter[ay] += "X"
				isEmpty = false
			} else {
				if _, haveDots := curLetter[ay]; haveDots {
					curLetter[ay] += " "
				}
			}
			ay++
		}

		if isEmpty {
			if len(curLetter) != 0 {
				if len(curAlpha) > 0 {
					curWidth-- // remove last blank column
					for yy, ln := range curLetter {
						if len(ln) >= curWidth {
							curLetter[yy] = ln[:curWidth]
						}
					}
					r, nbytes := utf8.DecodeRuneInString(curAlpha)
					allLetters[r] = curLetter
					curAlpha = curAlpha[nbytes:]
				}
				if curWidth > maxWidth {
					maxWidth = curWidth
				}
			}
			curWidth = 0
			curLetter = make(map[int]string)
		}
	}

	if *outName != "" {
		return
	}

	// output a simple text representation of the extracted characters
	for _, a := range *alphabet {
		if l, found := allLetters[a]; found {
			w := 0
			for yy := 0; yy < *height; yy++ {
				if len(l[yy]) > w {
					w = len(l[yy])
				}
			}

			leftPad := (maxWidth - w) / 2
			if *varWidth {
				leftPad = 0
			}
			for yy := 0; yy < *height; yy++ {
				l[yy] = strings.Repeat(" ", leftPad) + l[yy]
				fmt.Printf("%c  [%*s]\n", a, -maxWidth, l[yy])
			}
		}
	}
	return
}
