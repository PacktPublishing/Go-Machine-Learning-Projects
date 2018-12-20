package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"gocv.io/x/gocv"
)

var green = color.RGBA{0, 255, 0, 0}
var blue = color.RGBA{0, 0, 255, 0}

func main() {
	// open webcam
	webcam, err := gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	// classifier := gocvClassifier()
	// defer classifier.Close()

	// width := int(webcam.Get(gocv.VideoCaptureFrameWidth))
	// height := int(webcam.Get(gocv.VideoCaptureFrameHeight))

	// // set up pigo
	// goImg, grayGoImg, pigoClass, cParams, imgParams := pigoSetup(width, height)

	if ok := webcam.Read(&img); !ok {
		fmt.Printf("cannot read device %d\n", deviceID)
		return
	}
	if img.Empty() {
		log.Fatal("No image captures")
	}

	// log.Printf("img %v %v | %v %v", img.Rows(), img.Cols(), height, width)

	// if err = naughtyToImage(&img, goImg); err != nil {
	// 	log.Fatal(err)
	// }

	// grayGoImg = naughtyGrayscale(grayGoImg, goImg)
	// imgParams.Pixels = grayGoImg

	// // // detect faces
	// // rects := classifier.DetectMultiScale(img)
	// dets := pigoClass.RunCascade(imgParams, cParams)
	// dets = pigoClass.ClusterDetections(dets, 0.3)

	// for _, det := range dets {
	// 	if det.Q < 5 {
	// 		continue
	// 	}
	// 	x := det.Col - det.Scale/2
	// 	y := det.Row - det.Scale/2
	// 	r := image.Rect(x, y, x+det.Scale, y+det.Scale)
	// 	gocv.Rectangle(&img, r, green, 3)
	// 	size := gocv.GetTextSize("PIGO", gocv.FontHersheyPlain, 1.2, 2)
	// 	pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
	// 	gocv.PutText(&img, "PIGO", pt, gocv.FontHersheyPlain, 1.2, green, 2)
	// }
	goImg, err := img.ToImage()
	if err != nil {
		log.Fatal(err)
	}
	outFile, err := os.OpenFile("first.png", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	png.Encode(outFile, goImg)

	// mainloop(webcam)
}

func mainloop(webcam *gocv.VideoCapture) {
	var err error
	// open display window
	window := gocv.NewWindow("Face Detect")
	defer window.Close()

	// prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	// color for the rect when faces detected

	// load classifier to recognize faces
	classifier := gocvClassifier()
	defer classifier.Close()

	width := int(webcam.Get(gocv.VideoCaptureFrameWidth))
	height := int(webcam.Get(gocv.VideoCaptureFrameHeight))

	// set up pigo
	goImg, grayGoImg, pigoClass, cParams, imgParams := pigoSetup(width, height)

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("cannot read device %d\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		if err = naughtyToImage(&img, goImg); err != nil {
			log.Fatal(err)
		}
		grayGoImg = naughtyGrayscale(grayGoImg, goImg)
		imgParams.Pixels = grayGoImg

		// // detect faces
		rects := classifier.DetectMultiScale(img)
		dets := pigoClass.RunCascade(imgParams, cParams)
		dets = pigoClass.ClusterDetections(dets, 0.3)

		// draw a rectangle around each face on the original image,
		// along with text identifying as "Human"
		for _, r := range rects {
			gocv.Rectangle(&img, r, blue, 3)

			size := gocv.GetTextSize("GoCV", gocv.FontHersheyPlain, 1.2, 2)
			pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
			gocv.PutText(&img, "GoCV", pt, gocv.FontHersheyPlain, 1.2, blue, 2)
		}

		for _, det := range dets {
			if det.Q < 5 {
				continue
			}
			x := det.Col - det.Scale/2
			y := det.Row - det.Scale/2
			r := image.Rect(x, y, x+det.Scale, y+det.Scale)
			gocv.Rectangle(&img, r, green, 3)
			size := gocv.GetTextSize("PIGO", gocv.FontHersheyPlain, 1.2, 2)
			pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
			gocv.PutText(&img, "PIGO", pt, gocv.FontHersheyPlain, 1.2, green, 2)
		}

		// show the image in the window, and wait 1 millisecond
		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}
