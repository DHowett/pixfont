package parser

import "io"

type Matrix []uint32

type Font struct {
	Width, Height int
	Glyphs        map[rune]Matrix
}

type Parser interface {
	Decode(io.Reader) (*Font, error)
}
