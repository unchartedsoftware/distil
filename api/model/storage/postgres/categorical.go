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
	"math"
	"strconv"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil-compute/model"
	api "github.com/uncharted-distil/distil/api/model"
)

// CategoricalField defines behaviour for the categorical field type.
type CategoricalField struct {
	Storage     *Storage
	StorageName string
	Key         string
	Label       string
	Type        string
	subSelect   func() string
}

// NewCategoricalField creates a new field for categorical types.
func NewCategoricalField(storage *Storage, storageName string, key string, label string, typ string) *CategoricalField {
	field := &CategoricalField{
		Storage:     storage,
		StorageName: storageName,
		Key:         key,
		Label:       label,
		Type:        typ,
	}

	return field
}

// NewCategoricalFieldSubSelect creates a new field for categorical types
// and specifies a sub select query to pull the raw data.
func NewCategoricalFieldSubSelect(storage *Storage, storageName string, key string, label string, typ string, fieldSubSelect func() string) *CategoricalField {
	field := &CategoricalField{
		Storage:     storage,
		StorageName: storageName,
		Key:         key,
		Label:       label,
		Type:        typ,
		subSelect:   fieldSubSelect,
	}

	return field
}

// FetchSummaryData pulls summary data from the database and builds a histogram.
func (f *CategoricalField) FetchSummaryData(resultURI string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	var histogram *api.Histogram
	var err error
	if resultURI == "" {
		histogram, err = f.fetchHistogram(filterParams)
	} else {
		histogram, err = f.fetchHistogramByResult(resultURI, filterParams)
	}

	return histogram, err
}

func (f *CategoricalField) getTimeseriesAggQuery(extrema *api.Extrema) (string, string, string) {
	interval := extrema.GetBucketInterval()

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", api.HistogramAggPrefix, extrema.Key)
	rounded := extrema.GetBucketMinMax()

	bucketQueryString := ""
	// if only a single value, then return a simple count.
	if rounded.Max == rounded.Min {
		// want to return the count under bucket 0.
		bucketQueryString = fmt.Sprintf("(\"%s\" - \"%s\")", extrema.Key, extrema.Key)
	} else {
		bucketQueryString = fmt.Sprintf("width_bucket(\"%s\", %g, %g, %d) - 1",
			extrema.Key, rounded.Min, rounded.Max, extrema.GetBucketCount())
	}

	histogramQueryString := fmt.Sprintf("(%s) * %g + %g", bucketQueryString, interval, rounded.Min)

	return histogramAggName, bucketQueryString, histogramQueryString
}

func (f *CategoricalField) getTimeMinMaxAggsQuery(timeVar *model.Variable) string {
	// get min / max agg names
	minAggName := api.MinAggPrefix + timeVar.Name
	maxAggName := api.MaxAggPrefix + timeVar.Name

	timeSelect := fmt.Sprintf("CAST(\"%s\" AS INTEGER", timeVar.Name)
	if timeVar.Type == model.DateTimeType {
		timeSelect = fmt.Sprintf("CAST(extract(epoch from \"%s\") AS INTEGER)", timeVar.Name)
	}

	// create aggregations
	queryPart := fmt.Sprintf("MIN(%s) AS \"%s\", MAX(%s) AS \"%s\"",
		timeSelect, minAggName, timeSelect, maxAggName)
	// add aggregations
	return queryPart
}

func (f *CategoricalField) fetchTimeExtrema(timeVar *model.Variable) (*api.Extrema, error) {
	fromClause := f.getFromClause(true)

	// add min / max aggregation
	aggQuery := f.getTimeMinMaxAggsQuery(timeVar)

	// create a query that does min and max aggregations for each variable
	queryString := fmt.Sprintf("SELECT %s FROM %s;", aggQuery, fromClause)

	// execute the postgres query
	// NOTE: We may want to use the regular Query operation since QueryRow
	// hides db exceptions.
	res, err := f.Storage.client.Query(queryString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch extrema for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseTimeExtrema(timeVar, res)
}

func (f *CategoricalField) parseTimeExtrema(timeVar *model.Variable, rows *pgx.Rows) (*api.Extrema, error) {
	var minValue *int64
	var maxValue *int64
	if rows != nil {
		// Expect one row of data.
		exists := rows.Next()
		if !exists {
			return nil, fmt.Errorf("no rows in extrema query result")
		}
		err := rows.Scan(&minValue, &maxValue)
		if err != nil {
			return nil, errors.Wrap(err, "no min / max aggregation found")
		}
	}
	// check values exist
	if minValue == nil || maxValue == nil {
		return nil, errors.Errorf("no min / max aggregation values found")
	}
	// assign attributes
	return &api.Extrema{
		Key:  timeVar.Name,
		Type: timeVar.Type,
		Min:  float64(*minValue),
		Max:  float64(*maxValue),
	}, nil
}

func (f *CategoricalField) getTimeseriesHistogramAggQuery(extrema *api.Extrema) (string, string, string) {
	interval := extrema.GetBucketInterval()

	// get histogram agg name & query string.
	histogramAggName := fmt.Sprintf("\"%s%s\"", api.HistogramAggPrefix, extrema.Key)
	rounded := extrema.GetBucketMinMax()

	timeSelect := fmt.Sprintf("CAST(\"%s\" AS INTEGER", extrema.Key)
	if extrema.Type == model.DateTimeType {
		timeSelect = fmt.Sprintf("CAST(extract(epoch from \"%s\") AS INTEGER)", extrema.Key)
	}

	bucketQueryString := ""
	// if only a single value, then return a simple count.
	if rounded.Max == rounded.Min {
		// want to return the count under bucket 0.
		bucketQueryString = fmt.Sprintf("(%s - %s)", timeSelect, timeSelect)
	} else {
		bucketQueryString = fmt.Sprintf("width_bucket(%s, %g, %g, %d) - 1",
			timeSelect, rounded.Min, rounded.Max, extrema.GetBucketCount())
	}

	histogramQueryString := fmt.Sprintf("(%s) * %g + %g", bucketQueryString, interval, rounded.Min)

	return histogramAggName, bucketQueryString, histogramQueryString
}

func (f *CategoricalField) parseTimeHistogram(rows *pgx.Rows, extrema *api.Extrema) (*api.Histogram, error) {
	// get histogram agg name
	histogramAggName := api.HistogramAggPrefix + extrema.Key

	// Parse bucket results.
	interval := extrema.GetBucketInterval()

	buckets := make([]*api.Bucket, extrema.GetBucketCount())
	rounded := extrema.GetBucketMinMax()
	key := rounded.Min
	for i := 0; i < len(buckets); i++ {
		keyString := ""
		if model.IsFloatingPoint(extrema.Type) {
			keyString = fmt.Sprintf("%f", key)
		} else {
			keyString = strconv.Itoa(int(key))
		}

		buckets[i] = &api.Bucket{
			Key:   keyString,
			Count: 0,
		}

		key = key + interval
	}

	for rows.Next() {
		var bucketValue float64
		var bucketCount int64
		var bucket int64
		var category string
		err := rows.Scan(&bucket, &bucketValue, &category, &bucketCount)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("no %s histogram aggregation found", histogramAggName))
		}

		if bucket < 0 {
			// Due to float representation, sometimes the lowest value <
			// first bucket interval and so ends up in bucket -1.
			buckets[0].Count = bucketCount
		} else if bucket < int64(len(buckets)) {
			buckets[bucket].Count = bucketCount
		} else {
			// Since the max can match the limit, an extra bucket may exist.
			// Add the value to the second to last bucket.
			buckets[len(buckets)-1].Count += bucketCount
		}
	}
	// assign histogram attributes
	return &api.Histogram{
		Label:   f.Label,
		Key:     f.Key,
		Type:    model.NumericalType,
		VarType: f.Type,
		Extrema: rounded,
		Buckets: buckets,
	}, nil
}

func (f *CategoricalField) getTopCategories(filterParams *api.FilterParams) ([]string, error) {

	fromClause := f.getFromClause(true)

	// create the filter for the query
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, filterParams.Filters)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// get top N categories
	query := fmt.Sprintf("SELECT \"%s\", COUNT(*) AS count FROM %s %s GROUP BY \"%s\" ORDER BY count desc, \"%s\" LIMIT %d;",
		f.Key, fromClause, where, f.Key, f.Key, 5)

	rows, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if rows != nil {
		defer rows.Close()
	}

	var categories []string
	if rows != nil {
		for rows.Next() {
			var category string
			var count int64
			err := rows.Scan(&category, &count)
			if err != nil {
				return nil, err
			}
			categories = append(categories, category)
		}
	}
	return categories, nil
}

// FetchTimeseriesSummaryData pulls summary data from the database and builds a histogram.
func (f *CategoricalField) FetchTimeseriesSummaryData(timeVar *model.Variable, resultURI string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {

	if resultURI == "" {

		categories, err := f.getTopCategories(filterParams)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch top categories")
		}

		extrema, err := f.fetchTimeExtrema(timeVar)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch extrema from postgres")
		}

		histogramName, bucketQuery, histogramQuery := f.getTimeseriesHistogramAggQuery(extrema)

		// create the filter for the query.
		wheres := make([]string, 0)
		params := make([]interface{}, 0)
		wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, filterParams.Filters)

		categoryWhere := fmt.Sprintf("\"%s\" in (", f.Key)
		for index, category := range categories {
			categoryWhere += fmt.Sprintf("$%d", len(params)+1)
			if index < len(categories)-1 {
				categoryWhere += ","
			}
			params = append(params, category)
		}
		categoryWhere += ")"

		wheres = append(wheres, categoryWhere)
		where := fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))

		fromClause := f.getFromClause(true)

		// Create the complete query string.
		query := fmt.Sprintf("SELECT %s as bucket, CAST(%s as double precision) AS %s, \"%s\" as field, Count(*) AS count FROM %s %s GROUP BY %s, \"%s\" ORDER BY %s;",
			bucketQuery, histogramQuery, histogramName, f.Key, fromClause, where, bucketQuery, f.Key, histogramName)

		// execute the postgres query
		res, err := f.Storage.client.Query(query, params...)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
		}
		if res != nil {
			defer res.Close()
		}

		return f.parseTimeHistogram(res, extrema)
	}

	return nil, fmt.Errorf("not implemented")
}

func (f *CategoricalField) fetchHistogram(filterParams *api.FilterParams) (*api.Histogram, error) {
	fromClause := f.getFromClause(true)

	// create the filter for the query
	wheres := make([]string, 0)
	params := make([]interface{}, 0)
	wheres, params = f.Storage.buildFilteredQueryWhere(wheres, params, filterParams.Filters)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(wheres, " AND "))
	}

	// Get count by category.
	query := fmt.Sprintf("SELECT \"%s\", COUNT(*) AS count FROM %s %s GROUP BY \"%s\" ORDER BY count desc, \"%s\" LIMIT %d;",
		f.Key, fromClause, where, f.Key, f.Key, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res)
}

func (f *CategoricalField) fetchHistogramByResult(resultURI string, filterParams *api.FilterParams) (*api.Histogram, error) {
	fromClause := f.getFromClause(false)

	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.StorageName, resultURI, filterParams)
	if err != nil {
		return nil, err
	}

	params = append(params, resultURI)

	where := ""
	if len(wheres) > 0 {
		where = fmt.Sprintf("AND %s", strings.Join(wheres, " AND "))
	}

	// Get count by category.
	query := fmt.Sprintf(
		`SELECT data."%s", COUNT(*) AS count
		 FROM %s data INNER JOIN %s result ON data."%s" = result.index
		 WHERE result.result_id = $%d %s
		 GROUP BY "%s"
		 ORDER BY count desc, "%s" LIMIT %d;`,
		f.Key, fromClause, f.Storage.getResultTable(f.StorageName),
		model.D3MIndexFieldName, len(params), where, f.Key,
		f.Key, catResultLimit)

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for variable summaries from postgres")
	}
	if res != nil {
		defer res.Close()
	}

	return f.parseHistogram(res)
}

func (f *CategoricalField) parseHistogram(rows *pgx.Rows) (*api.Histogram, error) {
	termsAggName := api.TermsAggPrefix + f.Key

	// Parse bucket results.
	buckets := make([]*api.Bucket, 0)
	min := int64(math.MaxInt32)
	max := int64(-math.MaxInt32)

	if rows != nil {
		for rows.Next() {
			var term string
			var bucketCount int64
			err := rows.Scan(&term, &bucketCount)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("no %s histogram aggregation found", termsAggName))
			}

			buckets = append(buckets, &api.Bucket{
				Key:   term,
				Count: bucketCount,
			})
			if bucketCount < min {
				min = bucketCount
			}
			if bucketCount > max {
				max = bucketCount
			}
		}
	}

	// assign histogram attributes
	return &api.Histogram{
		Label:   f.Label,
		Key:     f.Key,
		Type:    model.CategoricalType,
		VarType: f.Type,
		Buckets: buckets,
		Extrema: &api.Extrema{
			Min: float64(min),
			Max: float64(max),
		},
	}, nil
}

// FetchPredictedSummaryData pulls predicted data from the result table and builds
// the categorical histogram for the field.
func (f *CategoricalField) FetchPredictedSummaryData(resultURI string, datasetResult string, filterParams *api.FilterParams, extrema *api.Extrema) (*api.Histogram, error) {
	targetName := f.Key

	// get filter where / params
	wheres, params, err := f.Storage.buildResultQueryFilters(f.StorageName, resultURI, filterParams)
	if err != nil {
		return nil, err
	}

	wheres = append(wheres, fmt.Sprintf("result.result_id = $%d AND result.target = $%d ", len(params)+1, len(params)+2))
	params = append(params, resultURI, targetName)

	query := fmt.Sprintf(
		`SELECT result.value, COUNT(*) AS count
		 FROM %s AS result INNER JOIN %s AS data ON result.index = data."%s"
		 WHERE %s
		 GROUP BY result.value
		 ORDER BY count desc;`,
		datasetResult, f.StorageName, model.D3MIndexFieldName, strings.Join(wheres, " AND "))

	// execute the postgres query
	res, err := f.Storage.client.Query(query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch histograms for result summaries from postgres")
	}
	defer res.Close()

	return f.parseHistogram(res)
}

func (f *CategoricalField) getFromClause(alias bool) string {
	fromClause := f.StorageName
	if f.subSelect != nil {
		fromClause = f.subSelect()
		if alias {
			fromClause = fmt.Sprintf("%s as %s", fromClause, f.StorageName)
		}
	}

	return fromClause
}
