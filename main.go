package main

import (
	"regexp"

	"github.com/fgrimme/ankiPDF/anki"
	"github.com/fgrimme/ankiPDF/config"
	"github.com/fgrimme/ankiPDF/document"
	"github.com/fgrimme/ankiPDF/layout"
	"github.com/jung-kurt/gofpdf"
	"github.com/microcosm-cc/bluemonday"
)

func main() {
	cfg, err := config.FromFile("./cfg.yaml")
	if err != nil {
		panic(err)
	}

	// we load the anki deck from file
	deck, err := anki.New("./anki-decks/01_NihongoShark.com__Kanji/01_NihongoShark.json")
	if err != nil {
		panic(err)
	}

	// cards have configured fields for front and back
	cards, err := anki.Cards(cfg, deck)
	if err != nil {
		panic(err)
	}

	// layout
	l := layout.New(cfg.CardSize)

	// document
	doc := document.New(l.PageSize, l.CardSize, cards)

	// render
	orientations := map[layout.Orientation]string{
		layout.Landscape: gofpdf.OrientationLandscape,
		layout.Portrait:  gofpdf.OrientationPortrait,
	}

	pdf := gofpdf.New(orientations[l.O], "mm", "A4", "./font")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)

	var x, y float64
	w := l.CardSize.W
	h := l.CardSize.H

	margin := cfg.Margin

	// default layout for front pages
	font := cfg.Front.Layout.Font
	size := cfg.Front.Layout.Size
	height := cfg.Front.Layout.Height
	align := cfg.Front.Layout.Align
	color := cfg.Front.Layout.Color

	// remove duplicate whitespaces
	space := regexp.MustCompile(`\s+`)

	// render front pages
	for _, page := range doc {
		pdf.AddPage()
		y = 0
		for _, row := range page {
			x = 0
			for _, card := range row {
				pdf.SetDrawColor(220, 220, 220)
				pdf.Rect(x, y, w, h, "D")
				pdf.SetXY(x+margin, y+margin)
				for _, field := range cfg.Front.Fields {
					// line-break
					if field == "break" {
						pdf.Ln(cfg.FieldLayouts["break"].Height)
						pdf.SetXY(x+margin, pdf.GetY())
						continue
					}
					// optional field fromatting from config
					if cfg.FieldLayouts[field].Size > 0 {
						size = cfg.FieldLayouts[field].Size
					}
					if cfg.FieldLayouts[field].Height > 0 {
						height = cfg.FieldLayouts[field].Height
					}
					if len(cfg.FieldLayouts[field].Font) > 0 {
						font = cfg.FieldLayouts[field].Font
					}
					if len(cfg.FieldLayouts[field].Align) > 0 {
						align = cfg.FieldLayouts[field].Align
					}
					if len(cfg.FieldLayouts[field].Color) > 0 {
						color = cfg.FieldLayouts[field].Color
					}
					// set formatting
					pdf.AddUTF8Font(font, "", font+".ttf")
					pdf.SetFont(font, "", size)
					pdf.SetTextColor(color[0], color[1], color[2])
					// render
					txt := card.Front[field]
					if cfg.StripHTML {
						txt = bluemonday.StrictPolicy().Sanitize(txt)
					}
					if cfg.TrimSpace {
						txt = space.ReplaceAllString(txt, " ")
					}
					pdf.MultiCell(w-2*margin, height, txt, "0", align, false)
					pdf.SetXY(x+margin, pdf.GetY())
				}
				x += w
			}
			y += h
		}
	}

	// default layout for back pages
	font = cfg.Back.Layout.Font
	size = cfg.Back.Layout.Size
	height = cfg.Back.Layout.Height
	align = cfg.Back.Layout.Align
	color = cfg.Back.Layout.Color

	// render back pages
	for _, page := range doc {
		pdf.AddPage()
		y = 0
		for _, row := range page {
			// draw from right to left
			x = l.PageSize.W - w
			// iterate cards in row from right to left
			for _, card := range row {
				pdf.SetDrawColor(220, 220, 220)
				pdf.Rect(x, y, w, h, "D")
				pdf.SetXY(x+margin, y+margin)

				for _, field := range cfg.Back.Fields {
					// optional field fromatting from config
					if cfg.FieldLayouts[field].Size > 0 {
						size = cfg.FieldLayouts[field].Size
					}
					if cfg.FieldLayouts[field].Height > 0 {
						height = cfg.FieldLayouts[field].Height
					}
					if len(cfg.FieldLayouts[field].Font) > 0 {
						font = cfg.FieldLayouts[field].Font
					}
					if len(cfg.FieldLayouts[field].Align) > 0 {
						align = cfg.FieldLayouts[field].Align
					}
					if len(cfg.FieldLayouts[field].Color) > 0 {
						color = cfg.FieldLayouts[field].Color
					}
					// set formatting
					pdf.AddUTF8Font(font, "", font+".ttf")
					pdf.SetFont(font, "", size)
					pdf.SetTextColor(color[0], color[1], color[2])
					// render
					txt := card.Back[field]
					if cfg.StripHTML {
						txt = bluemonday.StrictPolicy().Sanitize(txt)
					}
					if cfg.TrimSpace {
						txt = space.ReplaceAllString(txt, " ")
					}
					// render
					pdf.MultiCell(w-2*margin, height, txt, "0", align, false)
					pdf.SetXY(x+margin, pdf.GetY())
				}
				x -= w
			}
			y += h
		}
	}

	err = pdf.OutputFileAndClose("Fpdf_AddPage.pdf")
	if err != nil {
		panic(err)
	}

}
