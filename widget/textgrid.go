package widget

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
)

const (
	textAreaSpaceSymbol   = '·'
	textAreaTabSymbol     = '→'
	textAreaNewLineSymbol = '↵'
)

var (
	// TextGridStyleDefault is a default style for test grid cells
	TextGridStyleDefault TextGridStyle
	// TextGridStyleWhitespace is the style used for whitespace characters, if enabled
	TextGridStyleWhitespace TextGridStyle
)

// define the types seperately to the var definition so the custom style API is not leaked in their instances.
func init() {
	TextGridStyleDefault = &CustomTextGridStyle{}
	TextGridStyleWhitespace = &CustomTextGridStyle{FGColor: theme.ButtonColor()}
}

// TextGridCell represents a single cell in a text grid.
// It has a rune for the text content and a style associated with it.
type TextGridCell struct {
	Rune  rune
	Style TextGridStyle
}

// TextGridStyle defines a style that can be applied to a TextGrid cell.
type TextGridStyle interface {
	TextColor() color.Color
	BackgroundColor() color.Color
}

// CustomTextGridStyle is a utility type for those not wanting to define their own style types.
type CustomTextGridStyle struct {
	FGColor, BGColor color.Color
}

// TextColor is the color a cell should use for the text.
func (c *CustomTextGridStyle) TextColor() color.Color {
	return c.FGColor
}

// BackgroundColor is the color a cell should use for the background.
func (c *CustomTextGridStyle) BackgroundColor() color.Color {
	return c.BGColor
}

// TextGrid is a monospaced grid of characters.
// This is designed to be used by a text editor, code preview or terminal emulator.
type TextGrid struct {
	BaseWidget
	Content [][]TextGridCell

	LineNumbers bool
	Whitespace  bool
}

// MinSize returns the smallest size this widget can shrink to
func (t *TextGrid) MinSize() fyne.Size {
	t.ExtendBaseWidget(t)
	return t.BaseWidget.MinSize()
}

// Resize is called when this widget changes size. We should make sure that we refresh cells.
func (t *TextGrid) Resize(size fyne.Size) {
	t.BaseWidget.Resize(size)
	t.Refresh()
}

// SetText updates the buffer of this textgrid to contain the specified text.
// New lines and columns will be added as required. Lines are separated by '\n'.
// The grid will use default text style and any previous content and style will be removed.
func (t *TextGrid) SetText(text string) {
	rows := strings.Split(text, "\n")
	var buffer [][]TextGridCell
	for _, runes := range rows {
		var row []TextGridCell
		for _, r := range runes {
			row = append(row, TextGridCell{Rune: r})
		}
		buffer = append(buffer, row)
	}

	t.Content = buffer
	t.Refresh()
}

// Text returns the contents of the buffer as a single string (with no style information).
// It reconstructs the lines by joining with a `\n` character.
func (t *TextGrid) Text() string {
	ret := ""
	for i, row := range t.Content {
		for _, r := range row {
			ret += string(r.Rune)
		}

		if i < len(t.Content)-1 {
			ret += "\n"
		}
	}

	return ret
}

// Row returns the content of a specified row as a slice of TextGridCells.
// If the index is out of bounds it returns an empty slice.
func (t *TextGrid) Row(row int) []TextGridCell {
	if row < 0 || row >= len(t.Content) {
		return []TextGridCell{}
	}

	return t.Content[row]
}

// SetRow updates the specified row of the grid's contents using the specified cell content and style and then refreshes.
// If the row is beyond the end of the current buffer it will be expanded.
func (t *TextGrid) SetRow(row int, content []TextGridCell) {
	if row < 0 {
		return
	}
	for len(t.Content) <= row {
		t.Content = append(t.Content, []TextGridCell{})
	}

	t.Content[row] = content
	t.Refresh()
}

// SetStyle sets a grid style to the cell at named row and column
func (t *TextGrid) SetStyle(row, col int, style TextGridStyle) {
	if row < 0 || col < 0 {
		return
	}
	for len(t.Content) <= row {
		t.Content = append(t.Content, []TextGridCell{})
	}
	content := t.Content[row]

	for len(content) <= col {
		content = append(content, TextGridCell{})
	}
	content[col].Style = style
}

// SetStyleRange sets a grid style to all the cells between the start row and column through to the end row and column.
func (t *TextGrid) SetStyleRange(startRow, startCol, endRow, endCol int, style TextGridStyle) {
	if startRow == endRow {
		for col := startCol; col <= endCol; col++ {
			t.SetStyle(startRow, col, style)
		}
		return
	}

	// first row
	for col := startCol; col < len(t.Content[startRow]); col++ {
		t.SetStyle(startRow, col, style)
	}

	// possible middle rows
	for rowNum := startRow + 1; rowNum < endRow-1; rowNum++ {
		for col := 0; col < len(t.Content[rowNum]); col++ {
			t.SetStyle(rowNum, col, style)
		}
	}

	// last row
	for col := 0; col <= endCol; col++ {
		t.SetStyle(endRow, col, style)
	}
}

// CreateRenderer is a private method to Fyne which links this widget to it's renderer
func (t *TextGrid) CreateRenderer() fyne.WidgetRenderer {
	t.ExtendBaseWidget(t)
	render := &textGridRender{text: t}

	cell := canvas.NewText("M", color.White)
	cell.TextStyle.Monospace = true
	render.cellSize = cell.MinSize()

	return render
}

// NewTextGrid creates a new empty TextGrid widget.
func NewTextGrid() *TextGrid {
	grid := &TextGrid{}
	grid.ExtendBaseWidget(grid)
	return grid
}

// NewTextGridFromString creates a new TextGrid widget with the specified string content.
func NewTextGridFromString(content string) *TextGrid {
	grid := NewTextGrid()
	grid.SetText(content)
	return grid
}

type textGridRender struct {
	text *TextGrid

	cols, rows int

	cellSize fyne.Size
	objects  []fyne.CanvasObject
}

func (t *textGridRender) appendTextCell(str rune) {
	text := canvas.NewText(string(str), theme.TextColor())
	text.TextStyle.Monospace = true

	bg := canvas.NewRectangle(color.Transparent)
	t.objects = append(t.objects, bg, text)
}

func (t *textGridRender) setCellRune(str rune, pos int, style TextGridStyle) {
	rect := t.objects[pos*2].(*canvas.Rectangle)
	text := t.objects[pos*2+1].(*canvas.Text)
	if str == 0 {
		text.Text = " "
	} else {
		text.Text = string(str)
	}

	fg := theme.TextColor()
	if style != nil && style.TextColor() != nil {
		fg = style.TextColor()
	}
	text.Color = fg

	bg := color.Color(color.Transparent)
	if style != nil && style.BackgroundColor() != nil {
		bg = style.BackgroundColor()
	}
	rect.FillColor = bg
}

func (t *textGridRender) ensureGrid() {
	cellCount := t.cols * t.rows
	if len(t.objects) == cellCount*2 {
		return
	}
	for i := len(t.objects); i < cellCount*2; i += 2 {
		t.appendTextCell(' ')
	}
}

func (t *textGridRender) refreshGrid() {
	line := 1
	x := 0

	for rowIndex, row := range t.text.Content {
		if rowIndex >= t.rows { // would be an overflow - bad
			break
		}
		i := 0
		if t.text.LineNumbers {
			lineStr := []rune(fmt.Sprintf("%d", line))
			for c := 0; c < len(lineStr); c++ {
				t.setCellRune(lineStr[c], x, TextGridStyleWhitespace) // line numbers
				i++
				x++
			}
			for ; i < t.lineCountWidth(); i++ {
				t.setCellRune(' ', x, TextGridStyleWhitespace) // padding space
				x++
			}

			t.setCellRune('|', x, TextGridStyleWhitespace) // last space
			i++
			x++
		}
		for _, r := range row {
			if i >= t.cols { // would be an overflow - bad
				continue
			}
			if t.text.Whitespace && r.Rune == ' ' {
				if r.Style != nil && r.Style.BackgroundColor() != nil {
					whitespaceBG := &CustomTextGridStyle{FGColor: TextGridStyleWhitespace.TextColor(),
						BGColor: r.Style.BackgroundColor()}
					t.setCellRune(textAreaSpaceSymbol, x, whitespaceBG) // whitespace char
				} else {
					t.setCellRune(textAreaSpaceSymbol, x, TextGridStyleWhitespace) // whitespace char
				}
			} else {
				t.setCellRune(r.Rune, x, r.Style) // regular char
			}
			i++
			x++
		}
		if t.text.Whitespace && i < t.cols && rowIndex < len(t.text.Content)-1 {
			t.setCellRune(textAreaNewLineSymbol, x, TextGridStyleWhitespace) // newline
			i++
			x++
		}
		for ; i < t.cols; i++ {
			t.setCellRune(' ', x, TextGridStyleDefault) // blanks
			x++
		}

		line++
	}
	for ; x < len(t.objects)/2; x++ {
		t.setCellRune(' ', x, TextGridStyleDefault) // blank lines?
	}
	canvas.Refresh(t.text)
}

func (t *textGridRender) lineCountWidth() int {
	return len(fmt.Sprintf("%d", t.rows+1))
}

func (t *textGridRender) updateGridSize(size fyne.Size) {
	bufRows := len(t.text.Content)
	bufCols := 0
	for _, row := range t.text.Content {
		bufCols = int(math.Max(float64(bufCols), float64(len(row))))
	}
	sizeCols := int(math.Floor(float64(size.Width) / float64(t.cellSize.Width)))
	sizeRows := int(math.Floor(float64(size.Height) / float64(t.cellSize.Height)))

	if t.text.Whitespace {
		bufCols++
	}
	if t.text.LineNumbers {
		bufCols += t.lineCountWidth()
	}

	t.cols = int(math.Max(float64(sizeCols), float64(bufCols)))
	t.rows = int(math.Max(float64(sizeRows), float64(bufRows)))
}

func (t *textGridRender) Layout(size fyne.Size) {
	t.updateGridSize(size)
	t.ensureGrid()

	i := 0
	cellPos := fyne.NewPos(0, 0)
	for y := 0; y < t.rows; y++ {
		for x := 0; x < t.cols; x++ {
			t.objects[i*2+1].Move(cellPos)

			t.objects[i*2].Resize(t.cellSize)
			t.objects[i*2].Move(cellPos)
			cellPos.X += t.cellSize.Width
			i++
		}

		cellPos.X = 0
		cellPos.Y += t.cellSize.Height
	}
}

func (t *textGridRender) MinSize() fyne.Size {
	return fyne.NewSize(t.cellSize.Width*t.cols, t.cellSize.Height*t.rows)
}

func (t *textGridRender) Refresh() {
	t.ensureGrid()
	t.refreshGrid()
}

func (t *textGridRender) ApplyTheme() {
}

func (t *textGridRender) BackgroundColor() color.Color {
	return color.Transparent
}

func (t *textGridRender) Objects() []fyne.CanvasObject {
	return t.objects
}

func (t *textGridRender) Destroy() {
}
