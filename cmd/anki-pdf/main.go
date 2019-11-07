package main

import (
	"html"
	"path/filepath"
	"regexp"

	"github.com/fgrimme/anki-pdf/anki"
	"github.com/fgrimme/anki-pdf/config"
	"github.com/fgrimme/anki-pdf/document"
	"github.com/fgrimme/anki-pdf/layout"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/jung-kurt/gofpdf"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	version  = "unkown"
	cfgpath  = kingpin.Flag("cfg-path", "path to config file").Short('c').Required().String()
	ankipath = kingpin.Flag("anki-path", "path to anki deck JSON file").Short('a').Required().String()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	// deck specific configuration
	c, err := config.FromFile(*cfgpath)
	if err != nil {
		panic(err)
	}

	// load the anki deck from file
	deck, err := anki.New(*ankipath)
	if err != nil {
		panic(err)
	}

	// create cards from the anki deck
	cards, err := document.Cards(c, deck)
	if err != nil {
		panic(err)
	}

	// calc sizes and orientation
	l := layout.New(c.CardSize)

	// create a document with cards ordered in pages and rows
	d := document.New(l.PageSize, l.CardSize, cards)

	render(c, l, d)
}

func render(c *config.Config, l *layout.Layout, d document.Document) {
	orientations := map[layout.Orientation]string{
		layout.Landscape: gofpdf.OrientationLandscape,
		layout.Portrait:  gofpdf.OrientationPortrait,
	}
	pdf := gofpdf.New(orientations[l.O], "mm", "A4", "./font")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)

	// remove duplicate whitespaces
	space := regexp.MustCompile(`\s+`)

	// position on the page
	var x, y float64

	w := l.CardSize.W
	h := l.CardSize.H
	margin := c.Margin

	// render
	for _, page := range d {
		// front page
		pdf.AddPage()
		y = 0
		// default layout
		font := c.Front.Layout.Font
		size := c.Front.Layout.Size
		height := c.Front.Layout.Height
		align := c.Front.Layout.Align
		color := c.Front.Layout.Color
		for _, row := range page {
			x = 0
			for _, card := range row {
				pdf.SetDrawColor(220, 220, 220)
				pdf.Rect(x, y, w, h, "D")
				pdf.SetXY(x+margin, y+margin)
				for _, field := range c.Front.Fields {
					// line-break
					if field == "break" {
						pdf.Ln(c.FieldLayouts["break"].Height)
						pdf.SetXY(x+margin, pdf.GetY())
						continue
					}
					// optional field fromatting from config
					if c.FieldLayouts[field].Size > 0 {
						size = c.FieldLayouts[field].Size
					}
					if c.FieldLayouts[field].Height > 0 {
						height = c.FieldLayouts[field].Height
					}
					if len(c.FieldLayouts[field].Font) > 0 {
						font = c.FieldLayouts[field].Font
					}
					if len(c.FieldLayouts[field].Align) > 0 {
						align = c.FieldLayouts[field].Align
					}
					if len(c.FieldLayouts[field].Color) > 0 {
						color = c.FieldLayouts[field].Color
					}
					// set formatting
					pdf.AddUTF8Font(font, "", font+".ttf")
					pdf.SetFont(font, "", size)
					pdf.SetTextColor(color[0], color[1], color[2])
					// render
					txt := card.Front[field]
					if c.StripHTML {
						txt = strip.StripTags(txt)
					}
					if c.TrimSpace {
						txt = space.ReplaceAllString(txt, " ")
					}
					txt = html.UnescapeString(txt)
					pdf.MultiCell(w-2*margin, height, txt, "0", align, false)
					pdf.SetXY(x+margin, pdf.GetY())
				}
				x += w
			}
			y += h
		}
		// back page
		pdf.AddPage()
		y = 0
		// default layout
		font = c.Back.Layout.Font
		size = c.Back.Layout.Size
		height = c.Back.Layout.Height
		align = c.Back.Layout.Align
		color = c.Back.Layout.Color
		// render
		for _, row := range page {
			// draw from right to left
			x = l.PageSize.W - w
			// iterate cards in row from right to left
			for _, card := range row {
				pdf.SetDrawColor(220, 220, 220)
				pdf.Rect(x, y, w, h, "D")
				pdf.SetXY(x+margin, y+margin)

				for _, field := range c.Back.Fields {
					// optional field fromatting from config
					if c.FieldLayouts[field].Size > 0 {
						size = c.FieldLayouts[field].Size
					}
					if c.FieldLayouts[field].Height > 0 {
						height = c.FieldLayouts[field].Height
					}
					if len(c.FieldLayouts[field].Font) > 0 {
						font = c.FieldLayouts[field].Font
					}
					if len(c.FieldLayouts[field].Align) > 0 {
						align = c.FieldLayouts[field].Align
					}
					if len(c.FieldLayouts[field].Color) > 0 {
						color = c.FieldLayouts[field].Color
					}
					// set formatting
					pdf.AddUTF8Font(font, "", font+".ttf")
					pdf.SetFont(font, "", size)
					pdf.SetTextColor(color[0], color[1], color[2])
					// render
					txt := card.Back[field]
					if c.StripHTML {
						txt = strip.StripTags(txt)
					}
					if c.TrimSpace {
						txt = space.ReplaceAllString(txt, " ")
					}
					txt = html.UnescapeString(txt)
					pdf.MultiCell(w-2*margin, height, txt, "0", align, false)
					pdf.SetXY(x+margin, pdf.GetY())
				}
				x -= w
			}
			y += h
		}
	}

	outpath := *ankipath
	outpath = outpath[0 : len(outpath)-len(filepath.Ext(outpath))]
	err := pdf.OutputFileAndClose(outpath + ".pdf")
	if err != nil {
		panic(err)
	}

}
