// +build sidenote

package main

import (
	"image/color"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/gomono"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

var defaultFont vg.Font

func init() {
	font, err := truetype.Parse(gomono.TTF)
	if err != nil {
		panic(err)
	}
	vg.AddFont("gomono", font)
	defaultFont, err = vg.MakeFont("gomono", 12)
	if err != nil {
		panic(err)
	}
}

var table = []struct {
	x, m, c float64
}{
	{-0.9, -1.8, -0.803333333333333},
	{-0.8, -1.6, -0.633333333333334},
	{-0.7, -1.4, -0.483333333333334},
	{-0.6, -1.2, -0.353333333333333},
	{-0.5, -1, -0.243333333333333},
	{-0.4, -0.8, -0.153333333333333},
	{-0.3, -0.6, -0.083333333333333},
	{-0.2, -0.4, -0.033333333333333},
	{-0.1, -0.2, -0.003333333333333},
	{0, -2.71050543121376E-17, 0.006666666666667},
	{0.1, 0.2, -0.003333333333333},
	{0.2, 0.4, -0.033333333333333},
	{0.3, 0.6, -0.083333333333333},
	{0.4, 0.8, -0.153333333333333},
	{0.5, 1, -0.243333333333333},
	{0.6, 1.2, -0.353333333333333},
	{0.7, 1.4, -0.483333333333334},
	{0.8, 1.6, -0.633333333333333},
	{0.9, 1.8, -0.803333333333333},
}

type estimates []struct{ x, m, c float64 }

func (es estimates) Plot(c draw.Canvas, p *plot.Plot) {
	trX, trY := p.Transforms(&c)
	lineStyle := plotter.DefaultLineStyle
	lineStyle.Dashes = []vg.Length{vg.Points(2), vg.Points(2)}
	lineStyle.Color = color.RGBA{A: 255}
	for i, e := range es {
		if i == 0 || i == len(es)-1 {
			continue
		}
		strokeStartX := es[i-1].x
		strokeStartY := e.m*strokeStartX + e.c
		strokeEndX := es[i+1].x
		strokeEndY := e.m*strokeEndX + e.c
		x1 := trX(strokeStartX)
		y1 := trY(strokeStartY)
		x2 := trX(strokeEndX)
		y2 := trY(strokeEndY)
		x := trX(e.x)
		y := trY(e.x*e.m + e.c)

		c.DrawGlyph(plotter.DefaultGlyphStyle, vg.Point{X: x, Y: y})
		c.StrokeLine2(lineStyle, x1, y1, x2, y2)
	}
}

func main() {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "X^2 Function and Its Estimates"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"
	p.X.Min = -1.1
	p.X.Max = 1.1
	p.Y.Min = -0.1
	p.Y.Max = 1.1
	p.Y.Label.TextStyle.Font = defaultFont
	p.X.Label.TextStyle.Font = defaultFont
	p.X.Tick.Label.Font = defaultFont
	p.Y.Tick.Label.Font = defaultFont
	p.Title.Font = defaultFont
	p.Title.Font.Size = 16

	// Original function
	original := plotter.NewFunction(func(x float64) float64 { return x * x })
	original.Color = color.RGBA{A: 16}
	original.Width = 10
	p.Add(original)

	est := estimates(table)
	p.Add(est)

	if err := p.Save(25*vg.Centimeter, 25*vg.Centimeter, "functions.png"); err != nil {
		panic(err)
	}
}
