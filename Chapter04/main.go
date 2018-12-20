package main

import (
	"bufio"
	"image/color"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chewxy/stl"
	"github.com/jlaffaye/ftp"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

type loader func() io.Reader

func readFromFile() io.Reader {
	reader, err := os.Open("data.txt")
	dieIfErr(err)
	return reader
}

func download() io.Reader {
	client, err := ftp.Dial("aftp.cmdl.noaa.gov:21")
	dieIfErr(err)
	dieIfErr(client.Login("anonymous", "anonymous"))
	reader, err := client.Retr("products/trends/co2/co2_mm_mlo.txt")
	dieIfErr(err)
	return reader
}

func parse(l loader) (dates []string, co2s []float64) {
	s := bufio.NewScanner(l())
	for s.Scan() {
		row := s.Text()
		if strings.HasPrefix(row, "#") {
			continue
		}
		fields := strings.Fields(row)
		dates = append(dates, fields[2])
		co2, err := strconv.ParseFloat(fields[4], 64)
		dieIfErr(err)
		co2s = append(co2s, co2)
	}
	return
}

func main() {
	dateStrings, co2s := parse(readFromFile)
	dates := parseDates(dateStrings)
	plt := newTSPlot(dates, co2s, "CO2 Level")
	plt.X.Label.Text = "Time"
	plt.Y.Label.Text = "CO2 in the atmosphere (ppm)"
	plt.Title.Text = "CO2 in the atmosphere (ppm) over time\nTaken over the Mauna-Loa observatory"
	dieIfErr(plt.Save(25*vg.Centimeter, 25*vg.Centimeter, "Moana-Loa.png"))

	decomposed := stl.Decompose(co2s, 12, 84, stl.Additive(),
		stl.WithIter(1),
		// stl.WithTrendConfig(stl.Config{Jump: 1, Width: 18, Fn: loess.Linear}),
		// stl.WithSeasonalConfig(stl.Config{Jump: 1, Width: 1, Fn: loess.Linear}),
		// stl.WithLowpassConfig(stl.Config{Jump: 1, Width: 1, Fn: loess.Linear}),
	)

	dieIfErr(decomposed.Err)
	plts := plotDecomposed(dates, decomposed)
	writeToPng(plts, "CO2 in the atmosphere (ppm), decomposed", "decomposed.png", 25, 25)

	lies := stl.Decompose(co2s, 60, 84, stl.Additive(), stl.WithIter(1))
	dieIfErr(lies.Err)
	plts2 := plotDecomposed(dates, lies)
	writeToPng(plts2, "CO2 in the atmosphere (ppm), decomposed (Liar Edition)", "lies.png", 25, 25)

	fwd := 120
	forecast := hw(decomposed, 12, fwd, 0.1, 0.05, 0.1)
	datesplus := forecastTime(dates, fwd)
	forecastPlot := newTSPlot(datesplus, forecast, "")
	maxY := math.Inf(-1)
	minY := math.Inf(1)
	for i := range forecast {
		if forecast[i] > maxY {
			maxY = forecast[i]
		}
		if forecast[i] < minY {
			minY = forecast[i]
		}
	}
	// extend the range a little
	minY--
	maxY++
	maxX := float64(datesplus[len(datesplus)-1].Unix())
	minX := float64(datesplus[len(dates)-1].Unix())

	shadePoly := plotter.XYs{
		{X: minX, Y: minY},
		{X: maxX, Y: minY},
		{X: maxX, Y: maxY},
		{X: minX, Y: maxY},
	}
	poly, err := plotter.NewPolygon(shadePoly)
	dieIfErr(err)
	poly.Color = color.RGBA{A: 16}
	poly.LineStyle.Color = color.RGBA{}
	forecastPlot.Add(poly)

	writeToPng(forecastPlot, "Forecasted CO2 levels\n(10 years)", "forecast.png", 25, 25)
}

func dieIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func forecastTime(dates []time.Time, forwards int) []time.Time {
	retVal := append(dates, make([]time.Time, forwards)...)
	lastDate := dates[len(dates)-1]
	for i := len(dates); i < len(retVal); i++ {
		retVal[i] = lastDate.AddDate(0, 1, 0)
		lastDate = retVal[i]
	}
	return retVal
}

func hw(a stl.Result, periodicity, forward int, alpha, beta, gamma float64) []float64 {
	level := make([]float64, len(a.Data))
	trend := make([]float64, len(a.Trend))
	seasonal := make([]float64, len(a.Seasonal))
	forecast := make([]float64, len(a.Data)+forward)
	copy(seasonal, a.Seasonal)

	for i := range a.Data {
		if i == 0 {
			continue
		}
		level[i] = alpha*a.Data[i] + (1-alpha)*(level[i-1]+trend[i-1])
		trend[i] = beta*(level[i]-level[i-1]) + (1-beta)*(trend[i-1])
		if i-periodicity < 0 {
			continue
		}
		seasonal[i] = gamma*(a.Data[i]-level[i-1]-trend[i-1]) + (1-gamma)*(seasonal[i-periodicity])
	}

	hplus := ((periodicity - 1) % forward) + 1
	for i := 0; i+forward < len(forecast); i++ {
		forecast[i+forward] = level[i] + float64(forward)*trend[i] + seasonal[i-periodicity+hplus]
	}
	copy(forecast, a.Data)

	return forecast
}
