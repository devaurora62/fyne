package painter

import (
	"image"
	"image/draw"
	"log"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Drawer extends "golang.org/x/image/font" to add support for tabs
// Drawer draws text on a destination image.
//
// A Drawer is not safe for concurrent use by multiple goroutines, since its
// Face is not.
type Drawer struct {
	font.Drawer
}

func tabStop(f font.Face, x fixed.Int26_6) fixed.Int26_6 {
	spacew, ok := f.GlyphAdvance(' ')
	if !ok {
		log.Print("Failed to find space width for tab")
		return x
	}
	tabWidth := fyne.CurrentApp().Settings().Theme().Size(theme.SizeNameTabWidth)
	tabw := spacew * fixed.Int26_6(tabWidth)
	tabs, _ := math.Modf(float64((x + tabw) / tabw))
	return tabw * fixed.Int26_6(tabs)
}

// DrawString draws s at the dot and advances the dot's location.
// Tabs are translated into a dot location change.
func (d *Drawer) DrawString(s string) {
	prevC := rune(-1)
	for _, c := range s {
		if prevC >= 0 {
			d.Dot.X += d.Face.Kern(prevC, c)
		}
		if c == '\t' {
			d.Dot.X = tabStop(d.Face, d.Dot.X)
		} else {
			dr, mask, maskp, a, ok := d.Face.Glyph(d.Dot, c)
			if !ok {
				// TODO: is falling back on the U+FFFD glyph the responsibility of
				// the Drawer or the Face?
				// TODO: set prevC = '\ufffd'?
				continue
			}
			draw.DrawMask(d.Dst, dr, d.Src, image.Point{}, mask, maskp, draw.Over)
			d.Dot.X += a
		}

		prevC = c
	}
}

// MeasureString returns how far dot would advance by drawing s with f.
// Tabs are translated into a dot location change.
func MeasureString(f font.Face, s string) (advance fixed.Int26_6) {
	prevC := rune(-1)
	for _, c := range s {
		if prevC >= 0 {
			advance += f.Kern(prevC, c)
		}
		if c == '\t' {
			advance = tabStop(f, advance)
		} else {
			a, ok := f.GlyphAdvance(c)
			if !ok {
				// TODO: is falling back on the U+FFFD glyph the responsibility of
				// the Drawer or the Face?
				// TODO: set prevC = '\ufffd'?
				continue
			}
			advance += a
		}

		prevC = c
	}
	return advance
}
