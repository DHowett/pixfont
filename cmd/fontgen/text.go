package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

func processText(filename string) (allLetters map[rune]map[int]string, maxWidth int) {
	newalpha := ""
	input, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return nil, 0
	}
	allLetters = make(map[rune]map[int]string)
	re := regexp.MustCompile(`[^\n]*\n`)
	count := 0
	hh, maxHeight := 0, 0
	lastCh := rune(0)

	for _, bline := range re.FindAll(input, -1) {
		line := string(bline)
		c, pixoffs := utf8.DecodeRuneInString(line)
		pixoffs += 3
		if lastCh != c {
			count = 0
			hh = len(allLetters[lastCh])
			if hh > maxHeight {
				maxHeight = hh
			}
			allLetters[c] = make(map[int]string)
			newalpha += string(c)
		}
		ww := len(line) - (pixoffs + 2)
		if ww > maxWidth {
			maxWidth = ww
		}
		allLetters[c][count] = line[pixoffs : pixoffs+ww]
		lastCh = c
		count++
	}

	*alphabet = newalpha
	if *width == 0 {
		*width = maxWidth
	}
	if *height == 0 {
		*height = maxHeight
	}

	if *outName != "" {
		return
	}

	// output the same representation again, to allow user to verify it was parsed correctly
	for _, a := range *alphabet {
		if l, found := allLetters[a]; found {
			w := 0
			for yy := 0; yy < *height; yy++ {
				if len(l[yy]) > w {
					w = len(l[yy])
				}
			}

			leftPad := (maxWidth - w) / 2
			for yy := 0; yy < *height; yy++ {
				l[yy] = strings.Repeat(" ", leftPad) + l[yy]
				fmt.Printf("%c  [%*s]\n", a, -maxWidth, l[yy])
			}
		}
	}
	return
}
