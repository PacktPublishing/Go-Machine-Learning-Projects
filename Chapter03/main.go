package main

import (
	"fmt"
	"log"
)

func main() {
	typ := "lemm_stop"
	examples, err := ingest(typ)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Examples loaded: %d\n", len(examples))
	shuffle(examples)
	cvStart := len(examples) - len(examples)/3
	cv := examples[cvStart:]
	examples = examples[:cvStart]

	c := New()
	c.Train(examples)

	var corrects, totals float64
	for _, ex := range examples {
		// fmt.Printf("%v", c.Score(ham.Document))
		class := c.Predict(ex.Document)
		if class == ex.Class {
			corrects++
		}
		totals++
	}
	fmt.Printf("Dataset: %q. Corrects: %v, Totals: %v. Accuracy %v\n", typ, corrects, totals, corrects/totals)

	fmt.Println("Start Cross Validation (this classifier)")
	corrects, totals = 0, 0
	hams, spams := 0.0, 0.0
	var unseen, totalWords int
	for _, ex := range cv {
		totalWords += len(ex.Document)
		unseen += c.unseens(ex.Document)
		class := c.Predict(ex.Document)
		if class == ex.Class {
			corrects++
		}
		switch ex.Class {
		case Ham:
			hams++
		case Spam:
			spams++
		}
		totals++
	}

	fmt.Printf("Dataset: %q. Corrects: %v, Totals: %v. Accuracy %v\n", typ, corrects, totals, corrects/totals)
	fmt.Printf("Hams: %v, Spams: %v. Ratio to beat: %v\n", hams, spams, hams/(hams+spams))
	fmt.Printf("Previously unseen %d. Total Words %d\n", unseen, totalWords)
}
