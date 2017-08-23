package model

import (
	"gopkg.in/olivere/elastic.v5"

	"github.com/pkg/errors"
)

const (
	// MinAggPrefix is the prefix used for min aggregations.
	MinAggPrefix = "min_"
	// MaxAggPrefix is the prefix used for max aggregations.
	MaxAggPrefix = "max_"
	// TermsAggPrefix is the prefix used for terms aggregations.
	TermsAggPrefix = "terms_"
	// HistogramAggPrefix is the prefix used for histogram aggregations.
	HistogramAggPrefix = "histogram_"
	// VariableValueField is the field which stores the variable value.
	VariableValueField = "value"
	// VariableTypeField is the field which stores the variable's schema type value.
	VariableTypeField = "schemaType"
	// NumBuckets is the number of buckets to use for histograms
	NumBuckets = 50
)

// Extrema represents the extrema for a single variable.
type Extrema struct {
	Name string  `json:"-"`
	Type string  `json:"-"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

// Bucket represents a single histogram bucket.
type Bucket struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// Histogram represents a single variable histogram.
type Histogram struct {
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Extrema *Extrema  `json:"extrema,omitempty"`
	Buckets []*Bucket `json:"buckets"`
}

// FetchSummary returns the summary for the provided index, dataset, and
// variable.
func FetchSummary(storage Storage, client *elastic.Client, index string, dataset string, varName string) (*Histogram, error) {
	// need description of the variables to request aggregation against.
	variable, err := FetchVariable(client, index, dataset, varName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch variable description for summary")
	}

	return storage.FetchSummary(variable, dataset)
}
