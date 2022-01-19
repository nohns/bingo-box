package pdf

import (
	"fmt"
	"io"
	"strconv"

	bingo "github.com/nohns/bingo-box/server"
	"github.com/signintech/gopdf"
)

type CardFileGen struct {
	game           *bingo.Game
	pdf            gopdf.GoPdf
	currentPage    int
	margin         float64
	contentWidth   float64
	mHoriStart     float64
	mHoriEnd       float64
	mVertStart     float64
	mVertEnd       float64
	cardPad        float64
	cellLen        float64
	footerHeight   float64
	fontSize       float64
	numFontSize    float64
	footerFontSize float64
}

// Add a new page to the pdf file, and reset the y and x values.
func (fg *CardFileGen) newPage() {
	fg.pdf.AddPage()
	fg.currentPage++

	// Write page footer
	fg.setFontSize(fg.footerFontSize)
	fg.pdf.SetY(fg.mVertEnd - fg.footerHeight)
	fg.pdf.SetX(fg.mHoriStart)
	fg.centerTextInsideBounds(fmt.Sprintf("Page %d", fg.currentPage), fg.contentWidth, fg.footerHeight)

	// Reset x and y cursor
	fg.pdf.SetX(fg.mHoriStart)
	fg.pdf.SetY(fg.mVertStart)
}

// Draw a card grid to the pdf file
func (fg *CardFileGen) drawCard(card *bingo.Card) {
	baseY := fg.pdf.GetY()
	m := card.Matrix()

	// Setup
	fg.setFontSize(fg.numFontSize)

	// Print card
	for i := 0; i < 4; i++ {
		currY := baseY + float64(i)*fg.cellLen
		fg.pdf.SetX(fg.mHoriStart)
		fg.pdf.SetY(currY)
		fg.drawFullHoriLine()

		// Do not draw vertical lines when drawing bottom line
		if i >= 3 {
			continue
		}

		for j := 0; j < 10; j++ {
			fg.pdf.SetX(fg.mHoriStart + fg.cellLen*float64(j))
			fg.drawVertLine(fg.cellLen)

			// Do not draw numbers when drawing outer vertical line
			if j >= 9 {
				continue
			}

			// Render number to cell
			num := m[i][j]
			if num == 0 {
				continue
			}

			numStr := strconv.Itoa(num)
			fg.centerTextInsideBounds(numStr, fg.cellLen, fg.cellLen)

			// Reset y
			fg.pdf.SetY(currY)
		}
	}

	// Draw footer
	baseY = fg.pdf.GetY()
	fg.pdf.SetX(fg.mHoriEnd)
	fg.drawVertLine(fg.footerHeight)
	fg.pdf.SetX(fg.mHoriStart)
	fg.drawVertLine(fg.footerHeight)
	fg.pdf.SetX(fg.mHoriStart + 1)

	// Footer content left
	fg.setFontSize(fg.footerFontSize)
	fg.centerTextVertInHeight(fmt.Sprintf("Card number: %d", card.Number), fg.footerHeight)

	// Footer content center
	fg.pdf.SetX(fg.mHoriStart)
	fg.pdf.SetY(baseY)
	fg.centerTextInsideBounds(fg.game.Name, fg.contentWidth, fg.footerHeight)

	// Footer content right
	endCredit := "©	Bingo box 2021"
	ecLen, _ := fg.pdf.MeasureTextWidth(endCredit)
	fg.pdf.SetX(fg.mHoriEnd - ecLen - 1)
	fg.pdf.SetY(baseY)
	fg.centerTextVertInHeight(endCredit, fg.footerHeight)

	fg.pdf.SetY(baseY + fg.footerHeight)
	fg.drawFullHoriLine()

	// Add padding so next card can be printed from current y
	fg.pdf.SetY(fg.pdf.GetY() + fg.cardPad)
}

func (fg *CardFileGen) drawFullHoriLine() {
	currY := fg.pdf.GetY()
	fg.pdf.Line(fg.mHoriStart, currY, fg.mHoriEnd, currY)
}

func (fg *CardFileGen) drawVertLine(height float64) {
	currX := fg.pdf.GetX()
	currY := fg.pdf.GetY()
	fg.pdf.Line(currX, currY, currX, currY+height)
}

func (fg *CardFileGen) centerTextVertInHeight(text string, height float64) {
	fg.centerTextInsideBounds(text, 0, height)
}

func (fg *CardFileGen) centerTextHoriInWidth(text string, width float64) {
	fg.centerTextInsideBounds(text, width, 0)
}

func (fg *CardFileGen) centerTextInsideBounds(text string, width, height float64) {
	// default text position
	tx := fg.pdf.GetX()
	ty := fg.pdf.GetY()

	if width != 0 {
		tx = fg.calcTextInsideBoundsX(text, width)
	}
	if height != 0 {
		ty = fg.calcTextInsideBoundsY(height)
	}

	// Draw number as text in the correct position
	fg.pdf.SetX(tx)
	fg.pdf.SetY(ty)
	fg.pdf.Cell(nil, text)
}

func (fg *CardFileGen) calcTextInsideBoundsX(text string, width float64) float64 {
	txtWidth, _ := fg.pdf.MeasureTextWidth(text)
	// Get center
	cx := fg.pdf.GetX() + (width / 2)

	// Calculate position of num as it has to be in the middle of the cell
	return cx - (txtWidth / 2)
}

func (fg *CardFileGen) calcTextInsideBoundsY(height float64) float64 {
	// Get center
	cy := fg.pdf.GetY() + (height / 2)

	// Calculate y of text as it has to be in the middle of the height
	return cy - fg.fontSize/8
}

func (fg *CardFileGen) setFontSize(size float64) error {
	fg.fontSize = size
	return fg.pdf.SetFontSize(size)
}

func (fg *CardFileGen) Save(path string) error {
	return fg.pdf.WritePdf(path)
}

func (fg *CardFileGen) Write(w io.Writer) error {
	return fg.pdf.Write(w)
}

func GenFromCards(game *bingo.Game, cards []bingo.Card) (*CardFileGen, error) {

	// Create A4 pdf and by using mm as the unit
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4, Unit: gopdf.UnitMM})

	// Set default A4 margins
	margin := float64(15)
	pdf.SetMargins(margin, margin, margin, margin)

	// Add times font and use it by default
	err := pdf.AddTTFFont("Times New Roman", "pdf/font/times_new_roman.ttf")
	if err != nil {
		return nil, err
	}
	fontSize := float64(24)
	err = pdf.SetFont("Times New Roman", "", fontSize)
	if err != nil {
		return nil, err
	}

	// Set line width for cards
	pdf.SetLineWidth(0.5)

	// Create card file generator
	contentWidth := 210 - margin*2
	contentHeight := 297 - margin*2
	fg := &CardFileGen{
		game:           game,
		pdf:            pdf,
		margin:         margin,
		contentWidth:   contentWidth,
		mHoriStart:     margin,
		mHoriEnd:       margin + contentWidth,
		mVertStart:     margin,
		mVertEnd:       margin + contentHeight,
		cardPad:        30,
		cellLen:        contentWidth / 9,
		footerHeight:   5,
		fontSize:       fontSize,
		numFontSize:    24,
		footerFontSize: 10,
	}

	// Draw cards to file
	for i, c := range cards {
		// There can only be three cards on one page. Add a new one if necessary
		if i%3 == 0 {
			fg.newPage()
		}
		fg.drawCard(&c)
	}

	return fg, nil
}
