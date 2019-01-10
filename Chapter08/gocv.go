package main

import (
	"errors"
	"image"
	"image/color"
	"io/ioutil"
	"log"

	pigo "github.com/esimov/pigo/core"
	"gocv.io/x/gocv"
)

var haarCascadeFile string = "haarcascade_frontalface_default.xml"
var pigoCascadeFile []byte
var deviceID int = 0

func init() {
	var err error
	pigoCascadeFile, err = ioutil.ReadFile("facefinder")
	if err != nil {
		log.Fatalf("Error reading the cascade file: %v", err)
	}
}

func gocvClassifier() gocv.CascadeClassifier {
	// load classifier to recognize faces
	classifier := gocv.NewCascadeClassifier()
	if !classifier.Load(haarCascadeFile) {
		log.Fatalf("Error reading cascade file: %v\n", haarCascadeFile)
	}
	return classifier
}

func pigoSetup(width, height int) (*image.NRGBA, []uint8, *pigo.Pigo, pigo.CascadeParams, pigo.ImageParams) {
	goImg := image.NewNRGBA(image.Rect(0, 0, width, height))
	grayGoImg := make([]uint8, width*height)
	cParams := pigo.CascadeParams{
		MinSize:     20,
		MaxSize:     1000,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,
	}
	imgParams := pigo.ImageParams{
		Pixels: grayGoImg,
		Rows:   height,
		Cols:   width,
		Dim:    width,
	}
	pigo := pigo.NewPigo()
	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	classifier, err := pigo.Unpack(pigoCascadeFile)
	if err != nil {
		log.Fatalf("Error reading the cascade file: %s", err)
	}
	return goImg, grayGoImg, classifier, cParams, imgParams
}

func doOnePIGO(img *gocv.Mat, goImg *image.NRGBA, grayGoImg []uint8, pigoClass *pigo.Pigo, imgParams pigo.ImageParams, cParams pigo.CascadeParams) {
	var err error
	if err = naughtyToImage(img, goImg); err != nil {
		log.Fatal(err)
	}
	grayGoImg = naughtyGrayscale(grayGoImg, goImg)
	imgParams.Pixels = grayGoImg

	// // detect faces
	// rects := classifier.DetectMultiScale(img)
	dets := pigoClass.RunCascade(imgParams, cParams)
	dets = pigoClass.ClusterDetections(dets, 0.3)

}

func naughtyToImage(m *gocv.Mat, imge image.Image) error {
	typ := m.Type()
	if typ != gocv.MatTypeCV8UC1 && typ != gocv.MatTypeCV8UC3 && typ != gocv.MatTypeCV8UC4 {
		return errors.New("ToImage supports only MatType CV8UC1, CV8UC3 and CV8UC4")
	}

	width := m.Cols()
	height := m.Rows()
	step := m.Step()
	data := m.ToBytes()
	channels := m.Channels()

	switch img := imge.(type) {
	case *image.NRGBA:
		c := color.NRGBA{
			R: uint8(0),
			G: uint8(0),
			B: uint8(0),
			A: uint8(255),
		}
		for y := 0; y < height; y++ {
			for x := 0; x < step; x = x + channels {
				c.B = uint8(data[y*step+x])
				c.G = uint8(data[y*step+x+1])
				c.R = uint8(data[y*step+x+2])
				if channels == 4 {
					c.A = uint8(data[y*step+x+3])
				}
				img.SetNRGBA(int(x/channels), y, c)
			}
		}

	case *image.Gray:
		c := color.Gray{Y: uint8(0)}
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				c.Y = uint8(data[y*step+x])
				img.SetGray(x, y, c)
			}
		}
	}
	return nil
}

func naughtyGrayscale(dst []uint8, src *image.NRGBA) []uint8 {
	rows, cols := src.Bounds().Dx(), src.Bounds().Dy()
	if dst == nil || len(dst) != rows*cols {
		dst = make([]uint8, rows*cols)
	}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			dst[r*cols+c] = uint8(
				0.299*float64(src.Pix[r*4*cols+4*c+0]) +
					0.587*float64(src.Pix[r*4*cols+4*c+1]) +
					0.114*float64(src.Pix[r*4*cols+4*c+2]),
			)
		}
	}
	return dst
}
