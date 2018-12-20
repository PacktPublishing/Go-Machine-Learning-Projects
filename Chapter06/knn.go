package main

import (
	"sort"
)

func knn(a [][]float64, k int, distance func(a, b []float64) float64) ([][]float64, []float64) {
	var distances [][]float64
	for _, row := range a {
		var dists []float64
		for _, row2 := range a {
			dist := distance(row, row2)
			dists = append(dists, dist)
		}
		sort.Sort(sort.Float64Slice(dists))
		topK := dists[:k]
		distances = append(distances, topK)
	}

	var lastCol []float64
	for _, d := range distances {
		l := d[len(d)-1]
		lastCol = append(lastCol, l)
	}
	sort.Sort(sort.Float64Slice(lastCol))
	lastCol = sort.Reverse(sort.Float64Slice(lastCol)).(sort.Float64Slice)

	return distances, lastCol
}
