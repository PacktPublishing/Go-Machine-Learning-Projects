package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/ChimeraCoder/anaconda"
	"github.com/chewxy/lingo/corpus"
	"github.com/go-nlp/dmmclust"
	"github.com/go-nlp/tfidf"
	"github.com/mpraski/clusters"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type processedTweet struct {
	anaconda.Tweet

	// post processed stuff
	clean       string
	clean2      string
	ids         []int // to implement Document
	textVec     []float64
	normTextVec []float64
	location    []float64
	isRT        bool
}

func (t *processedTweet) IDs() []int { return t.ids }

func (t *processedTweet) Len() int { return len(t.ids) }

func toDocs(a []*processedTweet) []dmmclust.Document {
	retVal := make([]dmmclust.Document, len(a))
	for i := range a {
		retVal[i] = a[i]
	}
	return retVal
}

var nl = regexp.MustCompile("\n+")
var ht = regexp.MustCompile("&.+?;")

type processor struct {
	tfidf       *tfidf.TFIDF
	corpus      *corpus.Corpus
	transformer transform.Transformer
	locations   map[string]int
	locCount    int
}

func newProcessor() *processor {
	c, err := corpus.Construct(corpus.WithWords([]string{mention, hashtag, retweet, URL}))
	dieIfErr(err)

	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFKC)
	return &processor{
		tfidf:       tfidf.New(),
		corpus:      c,
		transformer: t,
		locations:   make(map[string]int),
	}
}

var xxx = 0

func (p *processor) single(word string) (wordID int, ok bool) {
	if _, ok = stopwords[word]; ok {
		return -1, false
	}
	switch {
	case strings.HasPrefix(word, "#"):
		word = strings.TrimPrefix(word, "#")
		// return p.corpus.Add(hashtag), true
	case strings.HasPrefix(word, "@"):
		if len(word) == 1 {
			return p.corpus.Add("at"), true
		}
		// return -1, false
		// return p.corpus.Add(mention), true
	case strings.HasPrefix(word, "http"):
		// return p.corpus.Add(URL), true
		return -1, false
	}

	if word == "rt" {
		return -1, false
		// return p.corpus.Add(retweet), false
	}

	return p.corpus.Add(word), true
}

func (p *processor) process(a []*processedTweet) []*processedTweet {
	// remove things from consideration
	i := 0
	for _, tt := range a {
		if tt.Lang == "en" {
			a[i] = tt
			i++
		}
	}
	a = a[:i]

	var err error
	for _, tt := range a {
		if tt.RetweetedStatus != nil {
			tt.Tweet = *tt.RetweetedStatus
		}

		tt.clean, _, err = transform.String(p.transformer, tt.FullText)
		dieIfErr(err)
		tt.clean = strings.ToLower(tt.clean)
		tt.clean = nl.ReplaceAllString(tt.clean, "\n")
		tt.clean = ht.ReplaceAllString(tt.clean, "")
		tt.clean = stripPunct(tt.clean)
		log.Printf("%v", tt.clean)
		for _, word := range strings.Fields(tt.clean) {
			// word = corpus.Singularize(word)
			wordID, ok := p.single(word)
			if ok {
				tt.ids = append(tt.ids, wordID)
				tt.clean2 += " "
				tt.clean2 += word
			}

			if word == "rt" {
				tt.isRT = true
			}
		}
		p.tfidf.Add(tt)
		log.Printf("%v", tt.clean2)
	}

	p.tfidf.CalculateIDF()
	// calculate scores
	for _, tt := range a {
		tt.textVec = p.tfidf.Score(tt)
	}

	// normalize text vector
	size := p.corpus.Size()
	for _, tt := range a {
		tt.normTextVec = make([]float64, size)
		for i := range tt.ids {
			tt.normTextVec[tt.ids[i]] = tt.textVec[i]
		}
	}
	return a
}

func stripPunct(a string) string {
	const punct = ",.?;:'\"!’*-“"
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(punct, r) < 0 {
			return r
		}
		return -1
	}, a)
}

func asMatrix(a []*processedTweet) [][]float64 {
	retVal := make([][]float64, len(a))
	for i := range a {
		retVal[i] = a[i].normTextVec
	}
	return retVal
}

func byClusters(a []int, expectedClusters int) (retVal [][]int) {
	if expectedClusters == 0 {
		return nil
	}
	retVal = make([][]int, expectedClusters)
	var i, v int
	defer func() {
		if r := recover(); r != nil {
			log.Printf("exp %v | %v", expectedClusters, v)
			panic(r)
		}
	}()
	for i, v = range a {
		if v == -1 {
			// retVal[0] = append(retVal[0], i)
			continue
		}
		retVal[v-1] = append(retVal[v-1], i)
	}
	return retVal
}

func byClusters2(a []dmmclust.Cluster, expectedClusters int) (retVal [][]int) {
	retVal = make([][]int, expectedClusters)
	for i, v := range a {
		retVal[v.ID()] = append(retVal[v.ID()], i)
	}
	return retVal
}

func load(r io.Reader) (retVal []*processedTweet) {
	dec := json.NewDecoder(r)
	dieIfErr(dec.Decode(&retVal))
	return retVal
}

func mock() []*processedTweet {
	f, err := os.Open("example.json")
	dieIfErr(err)
	return load(f)
}

func lift(ts []anaconda.Tweet) []*processedTweet {
	retVal := make([]*processedTweet, len(ts))
	for i := range ts {
		retVal[i] = new(processedTweet)
		retVal[i].Tweet = ts[i]
	}
	return retVal
}

func largestCluster(clusters []int) (int, int) {
	cc := make(map[int]int)
	for _, c := range clusters {
		cc[c]++
	}

	var retVal, maxVal int

	for k, v := range cc {
		if v > maxVal {
			retVal = k
			maxVal = v
		}
	}
	return retVal, cc[retVal]
}

func largestCluster2(clusters []dmmclust.Cluster) (int, int) {
	cc := make(map[int]int)
	for _, c := range clusters {
		cc[c.ID()]++
	}

	var retVal, maxVal int

	for k, v := range cc {
		if v > maxVal {
			retVal = k
			maxVal = v
		}
	}
	return retVal, cc[retVal]
}

func main() {
	f, err := os.Open("dev.json")
	dieIfErr(err)
	tweets := load(f)
	p := newProcessor()
	tweets = p.process(tweets)

	expC := 20
	distances, last := knn(asMatrix(tweets), expC, jaccard)
	log.Printf("distances %v | %v", distances, last)

	plt, err := plot.New()
	dieIfErr(err)
	plotutil.AddLinePoints(plt, "KNN Distance", plotKNNDist(last))
	plt.Save(25*vg.Centimeter, 25*vg.Centimeter, "KNNDist.png")

	dmmClust := dmm(tweets, expC, p.corpus.Size())
	kmeansClust := kmeans(tweets, expC)
	dbscanClust, clustCount := dbscan(tweets)

	// output
	log.Printf("len(tweets)%d", len(tweets))
	var buf bytes.Buffer

	bc := byClusters2(dmmClust, expC)
	lc, tweetCount := largestCluster2(dmmClust)
	fmt.Fprintf(&buf, "Largest Cluster %d - %d tweets\n", lc, tweetCount)
	for i, t := range bc {
		fmt.Fprintf(&buf, "CLUSTER %d: %d\n", i, len(t))
		for _, c := range t {
			fmt.Fprintf(&buf, "\t%v\n", tweets[c].clean2)
		}
	}
	fmt.Fprintf(&buf, "==============\n")
	bc2 := byClusters(kmeansClust, expC)
	for i, t := range bc2 {
		fmt.Fprintf(&buf, "CLUSTER %d: %d\n", i, len(t))
		for _, c := range t {
			fmt.Fprintf(&buf, "\t%v\n", tweets[c].clean2)
		}
	}
	fmt.Fprintf(&buf, "==============\n")
	bc3 := byClusters(dbscanClust, clustCount)
	for i, t := range bc3 {
		fmt.Fprintf(&buf, "CLUSTER %d: %d\n", i, len(t))
		for _, c := range t {
			fmt.Fprintf(&buf, "\t%v\n", tweets[c].clean2)
		}
	}

	log.Println(buf.String())

}

func plotKNNDist(a []float64) plotter.XYs {
	points := make(plotter.XYs, len(a))
	for i, val := range a {
		points[i].X = float64(i)
		points[i].Y = val
	}
	return points
}

func dmm(a []*processedTweet, expC int, corpusSize int) []dmmclust.Cluster {
	conf := dmmclust.Config{
		K:          expC,
		Vocabulary: corpusSize,
		Iter:       1000,
		Alpha:      0.0,
		Beta:       0.01,
		Score:      dmmclust.Algorithm4,
		Sampler:    dmmclust.NewGibbs(rand.New(rand.NewSource(1337))),
	}
	dmmClust, err := dmmclust.FindClusters(toDocs(a), conf)
	dieIfErr(err)
	return dmmClust
}

func kmeans(a []*processedTweet, expC int) []int {
	// create a clusterer
	// kmeans, err := clusters.KMeans(10000, 10, clusters.EuclideanDistance)
	kmeans, err := clusters.KMeans(100000, expC, jaccard)
	dieIfErr(err)
	data := asMatrix(a)
	dieIfErr(kmeans.Learn(data))
	return kmeans.Guesses()
}

func dbscan(a []*processedTweet) ([]int, int) {
	dbscan, err := clusters.DBSCAN(5, 0.965, 8, jaccard)
	dieIfErr(err)
	data := asMatrix(a)
	dieIfErr(dbscan.Learn(data))
	clust := dbscan.Guesses()

	counter := make(map[int]struct{})
	for _, c := range clust {
		counter[c] = struct{}{}
	}
	return clust, len(counter)
}

func jaccard(a, b []float64) float64 {

	setA, setB := make(map[int]struct{}), make(map[int]struct{})
	union := make(map[int]struct{})
	for i := range a {
		if a[i] != 0 {
			union[i] = struct{}{}
			setA[i] = struct{}{}
		}
	}

	for i := range b {
		if b[i] != 0 {
			union[i] = struct{}{}
			setB[i] = struct{}{}
		}
	}

	intersection := 0.0
	for k := range setA {
		if _, ok := setB[k]; ok {
			intersection++
		}
	}

	return 1 - (intersection / float64(len(union)))
}

/*
func main() {
	twitter := anaconda.NewTwitterApiWithCredentials(ACCESSTOKEN, ACCESSTOKENSECRET, CONSUMERKEY, CONSUMERSECRET)
	raw, err := twitter.GetHomeTimeline(url.Values{"count": []string{"200"}})
	// dieIfErr(err)
	// tweets := lift(raw)

	f, err := os.OpenFile("dev.json", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	dieIfErr(err)
	enc := json.NewEncoder(f)
	enc.Encode(raw)
	f.Close()
}
*/
func dieIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// isMn returns true if it's a non-spacing mark
func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}
