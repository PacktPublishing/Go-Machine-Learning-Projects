package main

import (
	"math"
	"math/rand"
	"sort"
	"strconv"
)

func tryNumCat(a string, index map[string][]int, catStrs []string) []string {
	isNumCat := true
	cats := make([]int, 0, len(index))
	for k := range index {
		i64, err := strconv.ParseInt(k, 10, 64)
		if err != nil && k != "NA" {
			isNumCat = false
			break
		}
		cats = append(cats, int(i64))
	}

	if isNumCat {
		sort.Ints(cats)
		for i := range cats {
			catStrs[i] = strconv.Itoa(cats[i])
		}
		if _, ok := index["NA"]; ok {
			catStrs[0] = "NA" // there are no negative numerical categories
		}
	} else {
		sort.Strings(catStrs)
	}
	return catStrs
}

func inList(a string, l []string) bool {
	for _, v := range l {
		if a == v {
			return true
		}
	}
	return false
}

// iqr finds the interquartile range
func iqr(a [][]float64, ql, qh float64, j int) (l, m, h float64) {
	var s []float64
	for _, row := range a {
		s = append(s, row[j])
	}
	sort.Float64s(s)
	idxLF64 := (ql / 100) * float64(len(s))
	idxHF64 := (qh / 100) * float64(len(s))

	// lower
	if math.Trunc(idxLF64) == idxLF64 {
		idxL := int(idxLF64)
		l = s[idxL-1]
	} else {
		idxL := int(idxLF64)
		l = (s[idxL] + s[idxL-1]) / float64(2)
	}

	// higher
	if math.Trunc(idxHF64) == idxHF64 {
		idxH := int(idxHF64)
		h = s[idxH-1]
	} else {
		idxH := int(idxHF64)
		h = (s[idxH] + s[idxH-1]) / float64(2)
	}

	// median
	if len(s)%2 == 0 {
		n := len(s)
		m = (s[n/2-1] + s[n/2+1]) / float64(2)
	} else {
		n := len(s)
		m = s[n/2]
	}

	return
}

func scale(a [][]float64, j int) {
	l, m, h := iqr(a, 0.25, 0.75, j)
	s := h - l
	if s == 0 {
		s = 1
	}

	for _, row := range a {
		row[j] = (row[j] - m) / s
	}
}

func scaleStd(a [][]float64, j int) {
	var mean, variance, n float64
	for _, row := range a {
		mean += row[j]
		n++
	}
	mean /= n
	for _, row := range a {
		variance += (row[j] - mean) * (row[j] - mean)
	}
	variance /= (n - 1)

	for _, row := range a {
		row[j] = (row[j] - mean) / variance
	}
}

func shuffle(a [][]float64, b []float64) {
	for i := len(a) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
		b[i], b[j] = b[j], b[i]
	}
}
