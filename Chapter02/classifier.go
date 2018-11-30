package main

import (
	"math"
	"math/rand"
	"sync"

	"github.com/chewxy/lingo/corpus"
	"github.com/go-nlp/tfidf"
)

const tiny = 0.0000001

type Class byte

const (
	Ham Class = iota
	Spam
	MAXCLASS
)

func (c Class) String() string {
	switch c {
	case Ham:
		return "Ham"
	case Spam:
		return "Spam"
	default:
		panic("HELP")
	}
}

// Example is a tuple representing a classification example
type Example struct {
	Document []string
	Class
}

type doc []int

func (d doc) IDs() []int { return []int(d) }

type Classifier struct {
	corpus *corpus.Corpus

	tfidfs [MAXCLASS]*tfidf.TFIDF
	totals [MAXCLASS]float64

	ready bool
	sync.Mutex
}

func New() *Classifier {
	var tfidfs [MAXCLASS]*tfidf.TFIDF
	for i := Ham; i < MAXCLASS; i++ {
		tfidfs[i] = tfidf.New()
	}
	return &Classifier{
		corpus: corpus.New(),
		tfidfs: tfidfs,
	}
}

func (c *Classifier) Train(examples []Example) {
	for _, ex := range examples {
		c.trainOne(ex)
	}
}

func (c *Classifier) Postprocess() {
	c.Lock()
	if c.ready {
		return
	}

	var docs int
	for _, t := range c.tfidfs {
		docs += t.Docs
	}
	for _, t := range c.tfidfs {
		t.Docs = docs
		// t.CalculateIDF()
		for k, v := range t.TF {
			t.IDF[k] = math.Log1p(float64(t.Docs) / v)
		}
	}
	c.ready = true
	c.Unlock()
}

func (c *Classifier) Score(sentence []string) (scores [MAXCLASS]float64) {
	if !c.ready {
		c.Postprocess()
	}

	d := make(doc, len(sentence))
	for i, word := range sentence {
		id := c.corpus.Add(word)
		d[i] = id
	}

	priors := c.priors()

	// score per class
	for i := range c.tfidfs {
		score := math.Log(priors[i])
		// likelihood
		for _, word := range sentence {
			prob := c.prob(word, Class(i))
			score += math.Log(prob)
		}

		scores[i] = score
	}
	return
}

func (c *Classifier) Predict(sentence []string) Class {
	scores := c.Score(sentence)
	return argmax(scores)
}

func (c *Classifier) unseens(sentence []string) (retVal int) {
	for _, word := range sentence {
		if _, ok := c.corpus.Id(word); !ok {
			retVal++
		}
	}
	return
}

func (c *Classifier) trainOne(example Example) {
	d := make(doc, len(example.Document))
	for i, word := range example.Document {
		id := c.corpus.Add(word)
		d[i] = id
	}
	c.tfidfs[example.Class].Add(d)
	c.totals[example.Class]++
}

func (c *Classifier) priors() (priors []float64) {
	priors = make([]float64, MAXCLASS)
	var sum float64
	for i, total := range c.totals {
		priors[i] = total
		sum += total
	}
	for i := Ham; i < MAXCLASS; i++ {
		priors[int(i)] /= sum
	}
	return
}

func (c *Classifier) prob(word string, class Class) float64 {
	id, ok := c.corpus.Id(word)
	if !ok {
		return tiny
	}

	freq := c.tfidfs[class].TF[id]
	idf := c.tfidfs[class].IDF[id]
	// idf := 1.0

	// a word may not appear at all in a class.
	if freq == 0 {
		return tiny
	}

	return freq * idf / c.totals[class]
}

func argmax(a [MAXCLASS]float64) Class {
	max := math.Inf(-1)
	var maxClass Class
	for i := Ham; i < MAXCLASS; i++ {
		score := a[i]
		if score > max {
			maxClass = i
			max = score
		}
	}
	return maxClass
}

func shuffle(a []Example) {
	// r := rand.New(rand.NewSource(time.Now().Unix()))

	for i := len(a) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
		// b[i], b[j] = b[j], b[i]
	}
	// for len(a) > 0 {
	// 	n := len(a)
	// 	randIndex := r.Intn(n)
	// 	a[n-1], a[randIndex] = a[randIndex], a[n-1]
	// 	a = a[:n-1]
	// }
}
