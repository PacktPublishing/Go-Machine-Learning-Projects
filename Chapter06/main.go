package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"gorgonia.org/tensor"
	"gorgonia.org/tensor/native"
)

func main() {
	imgs, err := readImageFile(os.Open("train-images-idx3-ubyte"))
	if err != nil {
		log.Fatal(err)
	}
	labels, err := readLabelFile(os.Open("train-labels-idx1-ubyte"))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("len imgs %d", len(imgs))
	data := prepareX(imgs)
	lbl := prepareY(labels)
	visualize(data, 10, 10, "image.png")
	data1, _ := data.Slice(makeRS(0, 1))
	log.Printf("%v", data1.Data())

	data2, err := zca(data)
	if err != nil {
		log.Fatal(err)
	}
	visualize(data2, 10, 10, "image2.png")
	data2x, _ := data2.Slice(makeRS(0, 1))
	log.Printf("%v", data2x.Data())
	_ = lbl

	nat, err := native.MatrixF64(data2.(*tensor.Dense))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Start Training")
	nn := New(784, 784, 10)
	costs := make([]float64, 0, data2.Shape()[0])
	for e := 0; e < 5; e++ {
		data2Shape := data2.Shape()
		var oneimg, onelabel tensor.Tensor
		for i := 0; i < data2Shape[0]; i++ {
			if oneimg, err = data.Slice(makeRS(i, i+1)); err != nil {
				log.Fatalf("Unable to slice one image %d", i)
			}
			if onelabel, err = lbl.Slice(makeRS(i, i+1)); err != nil {
				log.Fatalf("Unable to slice one label %d", i)
			}
			var cost float64
			if cost, err = nn.Train(oneimg, onelabel, 0.1); err != nil {
				log.Fatalf("Training error: %+v", err)
			}
			costs = append(costs, cost)
		}
		log.Printf("%d\t%v", e, avg(costs))
		shuffleX(nat)
		costs = costs[:0]
	}
	log.Printf("End training")

	log.Printf("Start testing")
	testImgs, err := readImageFile(os.Open("t10k-images.idx3-ubyte"))
	if err != nil {
		log.Fatal(err)
	}

	testlabels, err := readLabelFile(os.Open("t10k-labels.idx1-ubyte"))
	if err != nil {
		log.Fatal(err)
	}

	testData := prepareX(testImgs)
	testLbl := prepareY(testlabels)
	shape := testData.Shape()
	testData2, err := zca(testData)
	if err != nil {
		log.Fatal(err)
	}

	visualize(testData, 10, 10, "testData.png")
	visualize(testData2, 10, 10, "testData2.png")

	var correct, total float64
	var oneimg, onelabel tensor.Tensor
	var predicted, errcount int
	for i := 0; i < shape[0]; i++ {
		if oneimg, err = testData.Slice(makeRS(i, i+1)); err != nil {
			log.Fatalf("Unable to slice one image %d", i)
		}
		if onelabel, err = testLbl.Slice(makeRS(i, i+1)); err != nil {
			log.Fatalf("Unable to slice one label %d", i)
		}

		label := argmax(onelabel.Data().([]float64))
		if predicted, err = nn.Predict(oneimg); err != nil {
			log.Fatalf("Failed to predict %d", i)
		}

		if predicted == label {
			correct++
		} else if errcount < 5 {
			visualize(oneimg, 1, 1, fmt.Sprintf("%d_%d_%d.png", i, label, predicted))
			errcount++
		}
		total++
	}
	fmt.Printf("Correct/Totals: %v/%v = %1.3f\n", correct, total, correct/total)
	log.Printf("Correct/Totals: %v/%v = %1.3f\n", correct, total, correct/total)

	visualizeWeights(nn.hidden, 28, 28, "hiddenlayer.png")
	visualizeWeights(nn.final, 2, 5, "finallayer.png")

}

func shuffleX(a [][]float64) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tmp := make([]float64, len(a[0]))
	for i := range a {
		j := r.Intn(i + 1)
		copy(tmp, a[i])
		copy(a[i], a[j])
		copy(a[j], tmp)
	}
}

func argmax(a []float64) (retVal int) {
	var max = math.Inf(-1)
	for i := range a {
		if a[i] > max {
			retVal = i
			max = a[i]
		}
	}
	return
}

func avg(a []float64) (retVal float64) {
	s := sum(a)
	return s / float64(len(a))
}

func sum(a []float64) (retVal float64) {
	for i := range a {
		retVal += a[i]
	}
	return retVal
}
