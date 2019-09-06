//
//   Copyright © 2019 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package postgres

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// FetchResidualsExtremaByURI fetches the residual extrema by resultURI.
func (s *Storage) FetchResidualsExtremaByURI(dataset string, storageName string, resultURI string) (*api.Extrema, error) {
	storageNameResult := s.getResultTable(storageName)
	targetName, err := s.getResultTargetName(storageNameResult, resultURI)
	if err != nil {
		return nil, err
	}
	targetVariable, err := s.getResultTargetVariable(dataset, targetName)
	if err != nil {
		return nil, err
	}
	resultVariable := &model.Variable{
		Name: "value",
		Type: model.StringType,
	}
	return s.fetchResidualsExtrema(resultURI, storageName, targetVariable, resultVariable)
}

// FetchResidualsSummary fetches a histogram of the residuals associated with a set of numerical predictions.
func (s *Storage) FetchResidualsSummary(dataset string, storageName string, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.VariableSummary, error) {
	storageNameResult := s.getResultTable(storageName)
	targetName, err := s.getResultTargetName(storageNameResult, resultURI)
	if err != nil {
		return nil, err
	}

	variable, err := s.getResultTargetVariable(dataset, targetName)
	if err != nil {
		return nil, err
	}

	var baseline *api.Histogram
	var filtered *api.Histogram
	baseline, err = s.fetchResidualsSummary(dataset, storageName, variable, resultURI, nil, extrema)
	if err != nil {
		return nil, err
	}
	if !filterParams.Empty() {
		filtered, err = s.fetchResidualsSummary(dataset, storageName, variable, resultURI, filterParams, extrema)
		if err != nil {
			return nil, err
		}
	}

	return &api.VariableSummary{
		Label:    variable.DisplayName,
		Key:      variable.Name,
		Type:     model.NumericalType,
		VarType:  variable.Type,
		Baseline: baseline,
		Filtered: filtered,
	}, nil
}

func (s *Storage) fetchResidualsSummary(dataset string, storageName string, variable *model.Variable, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	// Just return a nil in the case where we were asked to return residuals for a non-numeric variable.
	if model.IsNumerical(variable.Type) || variable.Type == model.TimeSeriesType {
		// fetch numeric histograms
		residuals, err := s.fetchResidualsHistogram(resultURI, storageName, variable, filterParams, extrema)
		if err != nil {
			return nil, err
		}
		return residuals, nil
	}
	return nil, errors.Errorf("variable of type %s - should be numeric", variable.Type)
}

func getErrorTyped(variableName string) string {
	return fmt.Sprintf("(cast(value as double precision) - cast(\"%s\" as double precision))", variableName)
}

func (s *Storage) getResidualsHistogramAggQuery(extrema *api.Extrema, variableName string, resultVariable *model.Variable) (string, string, string) {
	// compute the bucket interval for the histogram
	interval := extrema.GetBucketInterval()

	// Only numeric types should occur.
	errorTyped := getErrorTyped(variableName)

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", api.HistogramAggPrefix, extrema.Key)
	rounded := extrema.GetBucketMinMax()
	bucketQueryString := fmt.Sprintf("width_bucket(%s, %g, %g, %d) - 1",
		errorTyped, rounded.Min, rounded.Max, extrema.GetBucketCount())
	histogramQueryString := fmt.Sprintf("(%s) * %g + %g", bucketQueryString, interval, rounded.Min)

	return histogramAggName, bucketQueryString, histogramQueryString
}

func getResultJoin(storageName string) string {
	// FROM clause to join result and base data on d3mIdex value
	return fmt.Sprintf("%s_result as res inner join %s as data on data.\"%s\" = res.index", storageName, storageName, model.D3MIndexFieldName)
}

func getResidualsMinMaxAggsQuery(variableName string, resultVariable *model.Variable) string {
	// get min / max agg names
	minAggName := api.MinAggPrefix + resultVariable.Name
	maxAggName := api.MaxAggPrefix + resultVariable.Name

	// Only numeric types should occur.
	errorTyped := getErrorTyped(variableName)

	// create aggregations
	queryPart := fmt.Sprintf("MIN(%s) AS \"%s\", MAX(%s) AS \"%s\"", errorTyped, minAggName, errorTyped, maxAggName)

	return queryPart
}

func (s *Storage) fetchResidualsExtrema(resultURI string, storageName string, variable *model.Variable,
	resultVariable *model.Variable) (*api.Extrema, error) {

	targetName := variable.Name
	if variable.Grouping != nil {
		targetName = variable.Grouping.Properties.YCol
	}

	// add min / max aggregation
	aggQuery := getResidualsMinMaxAggsQuery(targetName, resultVariable)

	// from clause to join result and base data
	fromClause := getResultJoin(storageName)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s WHERE result_id = $1 AND target = $2;", aggQuery, fromClause)

	// execute the postgres query
	res, err := s.client.Query(queryString, resultURI, variable.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for result from postgres")
	}
	defer res.Close()

	return s.parseExtrema(res, variable)
}

func (s *Storage) fetchResidualsHistogram(resultURI string, storageName string, variable *model.Variable, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	resultVariable := &model.Variable{
		Name: "value",
		Type: model.StringType,
	}

	targetName := variable.Name
	if variable.Grouping != nil {
		targetName = variable.Grouping.Properties.YCol
	}

	// need the extrema to calculate the histogram interval
	var err error
	if extrema == nil {
		extrema, err = s.fetchResidualsExtrema(resultURI, storageName, variable, resultVariable)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch result variable extrema for summary")
		}
	} else {
		extrema.Key = variable.Name
		extrema.Type = variable.Type
	}
	// for each returned aggregation, create a histogram aggregation. Bucket
	// size is derived from the min/max and desired bucket count.
	histogramName, bucketQuery, histogramQuery := s.getResidualsHistogramAggQuery(extrema, targetName, resultVariable)

	fromClause := getResultJoin(storageName)

	// create the filter for the query
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = s.buildFilteredQueryWhere(wheres, params, filterParams, false)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	// Create the complete query string.
	query := fmt.Sprintf(`
		SELECT %s as bucket, CAST(%s as double precision) AS %s, COUNT(*) AS count
		FROM %s
		WHERE result_id = $1 AND target = $2 %s
		GROUP BY %s ORDER BY %s;`, bucketQuery, histogramQuery, histogramName, fromClause, where, bucketQuery, histogramName)

	// execute the postgres query
	res, err := s.client.Query(query, resultURI, variable.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result variable summaries from postgres")
	}
	defer res.Close()

	field := NewNumericalField(s, storageName, variable.Name, variable.DisplayName, variable.Type)

	return field.parseHistogram(res, extrema)
}
