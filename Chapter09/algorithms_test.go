package main

import (
	"image"
	"testing"

	pigo "github.com/esimov/pigo/core"
	"gocv.io/x/gocv"
)

func BenchmarkGoCV(b *testing.B) {
	img := gocv.IMRead("test.png", gocv.IMReadUnchanged)
	if img.Cols() == 0 || img.Rows() == 0 {
		b.Fatalf("Unable to read image into file")
	}

	classifier := gocv.NewCascadeClassifier()
	if !classifier.Load(haarCascadeFile) {
		b.Fatalf("Error reading cascade file: %v\n", haarCascadeFile)
	}

	var rects []image.Rectangle
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rects = classifier.DetectMultiScale(img)
	}
	_ = rects
}

func BenchmarkPIGO(b *testing.B) {
	img := gocv.IMRead("test.png", gocv.IMReadUnchanged)
	if img.Cols() == 0 || img.Rows() == 0 {
		b.Fatalf("Unable to read image into file")
	}
	width := img.Cols()
	height := img.Rows()
	goImg, grayGoImg, pigoClass, cParams, imgParams := pigoSetup(width, height)

	var dets []pigo.Detection
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		grayGoImg = naughtyGrayscale(grayGoImg, goImg)
		imgParams.Pixels = grayGoImg
		dets = pigoClass.RunCascade(imgParams, cParams)
		dets = pigoClass.ClusterDetections(dets, 0.3)
	}
	_ = dets
}
