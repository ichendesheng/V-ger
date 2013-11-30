package shared

import (
	"time"
)

type AttributedString struct {
	Content string
	Style   int //0 -normal, 1 -italic, 2 -bold, 3 italic and bold
	Color   uint
}

type SubItem struct {
	From, To time.Duration
	Content  []AttributedString

	UsePosition bool
	Position
}
type Position struct {
	X, Y float64
}

type PlayProgressInfo struct {
	Left    string
	Right   string
	Percent float64
}