package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func parseDates(a []string) []time.Time {
	loc, err := time.LoadLocation("Pacific/Honolulu")
	dieIfErr(err)
	retVal := make([]time.Time, len(a))
	for i := range a {
		t, err := parseDecimalDate(a[i], loc)
		dieIfErr(err)
		retVal[i] = t
	}
	return retVal
}

// parseDecimalDate takes a string in format of a decimal date
//	"2018.05" and converts it into a date.
//
func parseDecimalDate(a string, loc *time.Location) (time.Time, error) {
	split := strings.Split(a, ".")
	if len(split) != 2 {
		return time.Time{}, errors.Errorf("Unable to split %q into a year followed by a decimal", a)
	}
	year, err := strconv.Atoi(split[0])
	if err != nil {
		return time.Time{}, err
	}
	dec, err := strconv.ParseFloat("0."+split[1], 64) // bugs can happen if you forget to add "0."
	if err != nil {
		return time.Time{}, err
	}

	// handle leap years
	var days float64 = 365
	if year%400 == 0 || year%4 == 0 && year%100 != 0 {
		days = 366
	}

	start := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	daysIntoYear := int(dec * days)
	retVal := start.AddDate(0, 0, daysIntoYear)
	return retVal, nil
}
