package main

import (
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
	"os"

	"gorgonia.org/tensor"
)

// Image holds the pixel intensities of an image.
// 255 is foreground (black), 0 is background (white).
type RawImage []byte

// Label is a digit label in 0 to 9
type Label uint8

const numLabels = 10
const pixelRange = 255

const (
	imageMagic = 0x00000803
	labelMagic = 0x00000801
	Width      = 28
	Height     = 28
)

var xxx int

func readLabelFile(r io.Reader, e error) (labels []Label, err error) {
	if e != nil {
		return nil, e
	}

	var magic, n int32
	if err = binary.Read(r, binary.BigEndian, &magic); err != nil {
		return nil, err
	}
	if magic != labelMagic {
		return nil, os.ErrInvalid
	}
	if err = binary.Read(r, binary.BigEndian, &n); err != nil {
		return nil, err
	}
	labels = make([]Label, n)
	for i := 0; i < int(n); i++ {
		var l Label
		if err := binary.Read(r, binary.BigEndian, &l); err != nil {
			return nil, err
		}
		labels[i] = l
	}
	return labels, nil
}

func readImageFile(r io.Reader, e error) (imgs []RawImage, err error) {
	if e != nil {
		return nil, e
	}

	var magic, n, nrow, ncol int32
	if err = binary.Read(r, binary.BigEndian, &magic); err != nil {
		return nil, err
	}
	if magic != imageMagic {
		return nil, err /*os.ErrInvalid*/
	}
	if err = binary.Read(r, binary.BigEndian, &n); err != nil {
		return nil, err
	}
	if err = binary.Read(r, binary.BigEndian, &nrow); err != nil {
		return nil, err
	}
	if err = binary.Read(r, binary.BigEndian, &ncol); err != nil {
		return nil, err
	}
	imgs = make([]RawImage, n)
	m := int(nrow * ncol)
	for i := 0; i < int(n); i++ {
		imgs[i] = make(RawImage, m)
		m_, err := io.ReadFull(r, imgs[i])
		if err != nil {
			return nil, err
		}
		if m_ != int(m) {
			return nil, os.ErrInvalid
		}
	}
	return imgs, nil
}

func pixelWeight(px byte) float64 {
	retVal := (float64(px) / 255 * 0.999) + 0.001
	if retVal == 1.0 {
		return 0.999
	}
	return retVal
}

func reversePixelWeight(px float64) byte {
	return byte(((px - 0.001) / 0.999) * 255)
}

func prepareX(M []RawImage) (retVal tensor.Tensor) {
	rows := len(M)
	cols := len(M[0])

	b := make([]float64, 0, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < len(M[i]); j++ {
			b = append(b, pixelWeight(M[i][j]))
		}
	}
	return tensor.New(tensor.WithShape(rows, cols), tensor.WithBacking(b))
}

func prepareY(N []Label) (retVal tensor.Tensor) {
	rows := len(N)
	cols := 10

	b := make([]float64, 0, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < 10; j++ {
			if j == int(N[i]) {
				b = append(b, 0.999)
			} else {
				b = append(b, 0.001)
			}
		}
	}
	return tensor.New(tensor.WithShape(rows, cols), tensor.WithBacking(b))
}

// visualize visualizes the first N images given a data tensor that is made up of float64s.
// It's arranged into (rows, 10) image.
// Row counts are calculated by dividing N by 10 - we only ever want 10 columns.
// For simplicity's sake, we will truncate any remainders.
func visualize(data tensor.Tensor, rows, cols int, filename string) (err error) {
	N := rows * cols

	sliced := data
	if N > 1 {
		sliced, err = data.Slice(makeRS(0, N), nil) // data[0:N, :] in python
		if err != nil {
			return err
		}
	}

	if err = sliced.Reshape(rows, cols, 28, 28); err != nil {
		return err
	}

	imCols := 28 * cols
	imRows := 28 * rows
	rect := image.Rect(0, 0, imCols, imRows)
	canvas := image.NewGray(rect)

	for i := 0; i < cols; i++ {
		for j := 0; j < rows; j++ {
			var patch tensor.Tensor
			if patch, err = sliced.Slice(makeRS(i, i+1), makeRS(j, j+1)); err != nil {
				return err
			}

			patchData := patch.Data().([]float64)
			for k, px := range patchData {
				x := j*28 + k%28
				y := i*28 + k/28
				c := color.Gray{reversePixelWeight(px)}
				canvas.Set(x, y, c)
			}
		}
	}

	var f io.WriteCloser
	if f, err = os.Create(filename); err != nil {
		return err
	}

	if err = png.Encode(f, canvas); err != nil {
		f.Close()
		return err
	}

	if err = f.Close(); err != nil {
		return err
	}
	return nil
}

func normalize(data tensor.Tensor) {
	raw := data.Data().([]float64)

	min, max := math.Inf(1), math.Inf(-1)
	for _, v := range raw {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}
	for i, v := range raw {
		raw[i] = v - min/(max-min)
	}
}

func visualizeWeights(w tensor.Tensor, rows, cols int, filename string) (err error) {

	s := w.Shape()
	log.Printf("s %v", s)

	var rehsapedTo []int
	var patchSide int
	switch s[1] {
	case 784:
		rehsapedTo = []int{rows, cols, 28, 28}
		patchSide = 28
	case 100:
		rehsapedTo = []int{rows, cols, 10, 10}
		patchSide = 10
	}

	sliced := w
	if err = sliced.Reshape(rehsapedTo...); err != nil {
		log.Printf("Err %v", err)
		return err
	}

	imCols := patchSide * cols
	imRows := patchSide * rows
	rect := image.Rect(0, 0, imCols, imRows)
	canvas := image.NewGray(rect)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			var patch tensor.Tensor
			if patch, err = sliced.Slice(makeRS(i, i+1), makeRS(j, j+1)); err != nil {
				log.Printf("FAILED TO SLICE i %d j %d ERR : %v | %v", i, j, err, sliced.Shape())
				return err
			}

			patchData := patch.Data().([]float64)
			for k, px := range patchData {
				x := j*patchSide + k%patchSide
				y := i*patchSide + k/patchSide
				c := color.Gray{reversePixelWeight(px)}
				canvas.Set(x, y, c)
			}
		}
	}

	var f io.WriteCloser
	if f, err = os.Create(filename); err != nil {
		log.Printf("FAILED TO CREATE FILE %v", filename)
		return err
	}

	if err = png.Encode(f, canvas); err != nil {
		log.Printf("FAILED TO ENCODE")
		f.Close()
		return err
	}

	if err = f.Close(); err != nil {
		return err
	}
	return nil
}
