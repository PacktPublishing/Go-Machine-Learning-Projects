package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type errList []error

func (err errList) Error() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Errors Found:\n")
	for _, e := range err {
		fmt.Fprintf(&buf, "\t%v\n", e)
	}
	return buf.String()
}

func ingest(typ string) (examples []Example, err error) {
	switch typ {
	case "bare", "lemm", "lemm_stop", "stop":
	default:
		return nil, errors.Errorf("Expected only \"bare\", \"lemm\", \"lemm_stop\" or \"stop\"")
	}

	var errs errList
	start, end := 0, 11
	// if isCrossVal {
	// 	start, end = 8, 11
	// }

	for i := start; i < end; i++ { // hold 30% for crossval
		matches, err := filepath.Glob(fmt.Sprintf("data/lingspam_public/%s/part%d/*.txt", typ, i))
		if err != nil {
			errs = append(errs, err)
			continue
		}

		for _, match := range matches {
			str, err := ingestOneFile(match)
			if err != nil {
				errs = append(errs, errors.WithMessage(err, match))
				continue
			}

			if strings.Contains(match, "spmsg") {
				// is spam
				// spams = append(spams, Example{str, Spam})
				examples = append(examples, Example{str, Spam})
			} else {
				// is ham
				// hams = append(hams, Example{str, Ham})
				examples = append(examples, Example{str, Ham})
			}
		}
	}
	if errs != nil {
		err = errs
	}
	return
}

func ingestOneFile(abspath string) ([]string, error) {
	bs, err := ioutil.ReadFile(abspath)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(bs), " "), nil
}
