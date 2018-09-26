package postgres

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/model"
)

// VectorField defines behaviour for any Vector type.
type VectorField struct {
	Storage *Storage
}

// NewVectorField creates a new field of the vector type. A vector field
// uses unnest to flatten the database array and then uses the underlying
// data type to get summaries.
func NewVectorField(storage *Storage) *VectorField {
	field := &VectorField{
		Storage: storage,
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *VectorField) FetchSummaryData(dataset string, variable *model.Variable, resultURI string, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	var underlyingField Field
	if f.isNumerical(variable) {
		underlyingField = NewNumericalFieldSubSelect(f.Storage, f.subSelect)
	} else {
		underlyingField = NewCategoricalFieldSubSelect(f.Storage, f.subSelect)
	}

	return underlyingField.FetchSummaryData(dataset, variable, resultURI, filterParams, extrema)
}

// FetchNumericalStats gets the variable's numerical summary info (mean, stddev).
func (f *VectorField) FetchNumericalStats(dataset string, variable *model.Variable, filterParams *model.FilterParams) (*NumericalStats, error) {
	// confirm that the underlying type is numerical
	if !f.isNumerical(variable) {
		return nil, errors.Errorf("field '%s' is not a numerical vector", variable.Key)
	}

	// use the underlying numerical field implementation
	field := NewNumericalFieldSubSelect(f.Storage, f.subSelect)

	return field.FetchNumericalStats(dataset, variable, filterParams)
}

// FetchNumericalStatsByResult gets the variable's numerical summary info (mean, stddev) for a result set.
func (f *VectorField) FetchNumericalStatsByResult(dataset string, variable *model.Variable, resultURI string, filterParams *model.FilterParams) (*NumericalStats, error) {
	// confirm that the underlying type is numerical
	if !f.isNumerical(variable) {
		return nil, errors.Errorf("field '%s' is not a numerical vector", variable.Key)
	}

	// use the underlying numerical field implementation
	field := NewNumericalFieldSubSelect(f.Storage, f.subSelect)

	return field.FetchNumericalStatsByResult(dataset, variable, resultURI, filterParams)
}

// FetchPredictedSummaryData pulls predicted data from the result table and builds
// the categorical histogram for the field.
func (f *VectorField) FetchPredictedSummaryData(resultURI string, dataset string, datasetResult string, variable *model.Variable, filterParams *model.FilterParams, extrema *model.Extrema) (*model.Histogram, error) {
	return nil, errors.Errorf("vector field cannot be a target so no result will be pulled")
}

func (f *VectorField) isNumerical(variable *model.Variable) bool {
	return model.IsNumerical(strings.Replace(variable.Type, "Vector", "", -1))
}

func (f *VectorField) subSelect(dataset string, variable *model.Variable) string {
	return fmt.Sprintf("(SELECT %s, unnest(\"%s\") as %s FROM %s)",
		model.D3MIndexFieldName, variable.Key, variable.Key, dataset)
}
