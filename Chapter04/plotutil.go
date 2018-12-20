package main

import (
	"image/color"
	"math"
	"os"
	"time"

	"golang.org/x/image/font/gofont/gomono"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"

	"github.com/chewxy/stl"
	"github.com/golang/freetype/truetype"
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

type dateTicks []time.Time

func (t dateTicks) Ticks(min, max float64) []plot.Tick {
	var retVal []plot.Tick
	for i := math.Trunc(min); i <= max; i++ {
		retVal = append(retVal, plot.Tick{Value: i, Label: t[int(i)].String()})
	}
	return retVal
}

type residChart struct {
	plotter.XYs
	draw.LineStyle
}

func (r *residChart) Plot(c draw.Canvas, p *plot.Plot) {
	xmin, xmax, ymin, ymax := r.DataRange()
	p.Y.Min = ymin
	p.Y.Max = ymax
	p.X.Min = xmin
	p.X.Max = xmax

	trX, trY := p.Transforms(&c)
	zero := trY(0)
	lineStyle := r.LineStyle
	for _, xy := range r.XYs {
		x := trX(xy.X)
		y := trY(xy.Y)
		c.StrokeLine2(lineStyle, x, zero, x, y)
	}
}

func (r *residChart) DataRange() (xmin, xmax, ymin, ymax float64) {
	xmin = math.Inf(1)
	xmax = math.Inf(-1)
	ymin = math.Inf(1)
	ymax = math.Inf(-1)
	for _, xy := range r.XYs {
		xmin = math.Min(xmin, xy.X)
		xmax = math.Max(xmax, xy.X)
		ymin = math.Min(ymin, xy.Y)
		ymax = math.Max(ymax, xy.Y)
	}
	return
}

func (r *residChart) Thumbnail(c *draw.Canvas) {
	y := c.Center().Y
	c.StrokeLine2(r.LineStyle, c.Min.X, y, c.Max.X, y)
}

func newTSPlot(xs []time.Time, ys []float64, seriesName string) *plot.Plot {
	p, err := plot.New()
	dieIfErr(err)
	xys := make(plotter.XYs, len(ys))
	for i := range ys {
		xys[i].X = float64(xs[i].Unix())
		xys[i].Y = ys[i]
	}
	l, err := plotter.NewLine(xys)
	dieIfErr(err)
	l.LineStyle.Color = color.RGBA{A: 255} // black
	p.Add(l)
	if seriesName != "" {
		p.Legend.Add(seriesName, l)
		p.Legend.TextStyle.Font = defaultFont
	}

	// dieIfErr(plotutil.AddLines(p, seriesName, xys))
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-01"}
	p.Y.Label.TextStyle.Font = defaultFont
	p.X.Label.TextStyle.Font = defaultFont
	p.X.Tick.Label.Font = defaultFont
	p.Y.Tick.Label.Font = defaultFont
	p.Title.Font = defaultFont
	p.Title.Font.Size = 16

	return p
}

func newResidPlot(xs []time.Time, ys []float64, seriesName string) *plot.Plot {
	p, err := plot.New()
	dieIfErr(err)
	xys := make(plotter.XYs, len(ys))
	for i := range ys {
		xys[i].X = float64(xs[i].Unix())
		xys[i].Y = ys[i]
	}
	r := &residChart{XYs: xys, LineStyle: plotter.DefaultLineStyle}
	r.LineStyle.Color = color.RGBA{A: 255}
	p.Add(r)
	p.Legend.Add(seriesName, r)

	p.Legend.TextStyle.Font = defaultFont
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-01"}
	p.Y.Label.TextStyle.Font = defaultFont
	p.X.Label.TextStyle.Font = defaultFont
	p.X.Tick.Label.Font = defaultFont
	p.Y.Tick.Label.Font = defaultFont
	p.Title.Font.Size = 16
	return p
}

func plotDecomposed(xs []time.Time, a stl.Result) [][]*plot.Plot {
	plots := make([][]*plot.Plot, 4)
	plots[0] = []*plot.Plot{newTSPlot(xs, a.Data, "Data")}
	plots[1] = []*plot.Plot{newTSPlot(xs, a.Trend, "Trend")}
	plots[2] = []*plot.Plot{newTSPlot(xs, a.Seasonal, "Seasonal")}
	plots[3] = []*plot.Plot{newResidPlot(xs, a.Resid, "Residuals")}

	return plots
}

func writeToPng(a interface{}, title, filename string, width, height vg.Length) {
	switch at := a.(type) {
	case *plot.Plot:
		at.Title.Text = title
		dieIfErr(at.Save(width*vg.Centimeter, height*vg.Centimeter, filename))
		return
	case [][]*plot.Plot:
		rows := len(at)
		cols := len(at[0])
		t := draw.Tiles{
			Rows: rows,
			Cols: cols,
		}
		img := vgimg.New(width*vg.Centimeter, height*vg.Centimeter)
		dc := draw.New(img)

		if title != "" {
			at[0][0].Title.Text = title
		}

		canvases := plot.Align(at, t, dc)
		for i := 0; i < t.Rows; i++ {
			for j := 0; j < t.Cols; j++ {
				at[i][j].Draw(canvases[i][j])
			}
		}

		w, err := os.Create(filename)
		dieIfErr(err)

		png := vgimg.PngCanvas{Canvas: img}
		_, err = png.WriteTo(w)
		dieIfErr(err)
		return
	}
	panic("Unreachable")

}
