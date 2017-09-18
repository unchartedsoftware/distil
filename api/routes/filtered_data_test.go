package routes

import (
	"net/http"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/unchartedsoftware/distil/api/model/storage/elastic"
	"github.com/unchartedsoftware/distil/api/model/storage/postgres"
	pg "github.com/unchartedsoftware/distil/api/postgres"
	"github.com/unchartedsoftware/distil/api/util/json"
	"github.com/unchartedsoftware/distil/api/util/mock"
)

func TestParseSearchParams(t *testing.T) {
	params := map[string]string{
		"dataset": "o_185",
	}
	query := map[string]string{
		"On_base_pct": "numerical,0,100",
		"Position":    "categorical,Catcher,Pitcher",
		"Triples":     "",
		"size":        "100",
	}
	req := mock.HTTPRequest(t, "GET", "/distil/data", params, query)
	// parse
	filters, err := parseFilterParams(req)
	assert.NoError(t, err)
	assert.Equal(t, filters.Ranged[0].Name, "On_base_pct")
	assert.Equal(t, filters.Ranged[0].Min, 0.0)
	assert.Equal(t, filters.Ranged[0].Max, 100.0)
	assert.Equal(t, filters.Categorical[0].Name, "Position")
	assert.Equal(t, filters.Categorical[0].Categories, []string{"Catcher", "Pitcher"})
	assert.Equal(t, filters.None[0], "Triples")
	assert.Equal(t, filters.Size, 100)
}

func TestParseSearchParamsMalformed(t *testing.T) {
	params := map[string]string{
		"dataset": "o_185",
	}
	query0 := map[string]string{
		"On_base_pct": "numerical,0",
	}
	query1 := map[string]string{
		"Position": "categorical",
	}
	query2 := map[string]string{
		"Triples": "numerical,1,2,3",
	}
	query3 := map[string]string{
		"size": "",
	}
	// missing max
	req := mock.HTTPRequest(t, "GET", "/distil/data", params, query0)
	filters, err := parseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
	// missing categories
	req = mock.HTTPRequest(t, "GET", "/distil/data", params, query1)
	filters, err = parseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
	// extra param
	req = mock.HTTPRequest(t, "GET", "/distil/data", params, query2)
	filters, err = parseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
	// missing size
	req = mock.HTTPRequest(t, "GET", "/distil/data", params, query3)
	filters, err = parseFilterParams(req)
	assert.Error(t, err)
	assert.Nil(t, filters)
}

func TestFilteredDataHandler(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{"./testdata/filtered_data.json"})
	// mock elasticsearch client
	ctor := mock.ElasticClientCtor(t, handler)

	// instantiate storage client constructor.
	storageCtor := elastic.NewStorage(ctor)

	// put together a stub dataset request
	params := map[string]string{
		"dataset": "o_185",
		"esIndex": "dataset",
	}
	query := map[string]string{
		"On_base_pct": "numerical,0,100",
		"Position":    "categorical,Catcher",
		"Triples":     "",
	}
	req := mock.HTTPRequest(t, "GET", "/distil/data", params, query)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, FilteredDataHandler(storageCtor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
			"name": "o_185",
			"metadata": [
				{"name": "On_base_pct", "type": "float", "importance": 0},
				{"name": "Position", "type": "categorical", "importance": 0},
				{"name": "Triples", "type": "integer", "importance": 0}
			],
			"values": [
				[0.268, "Catcher", 5],
				[0.244, "Catcher", 1]
			]
		}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFilteredPostgresHandler(t *testing.T) {
	// mock elasticsearch request handler
	handler := mock.ElasticHandler(t, []string{"./testdata/filtered_data.json"})
	// mock elasticsearch client
	ctorES := mock.ElasticClientCtor(t, handler)

	// mock postgres client
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDB := mock.NewDatabaseDriver(ctrl)

	ctor := mockContructor(mockDB)

	// instantiate storage filter client constructor.
	storageCtor := postgres.NewStorage(ctor, ctorES)

	// put together a stub dataset request
	params := map[string]string{
		"dataset": "o_185",
		"esIndex": "dataset",
	}
	query := map[string]string{
		"On_base_pct": "numerical,0,100",
		"Position":    "categorical,Catcher",
		"Triples":     "",
	}

	// Identify the expected behaviour.
	// NOTE: It currently expects an empty set since pgx.Rows is hardly accessible.
	mockDB.EXPECT().Query(
		"SELECT  FROM o_185 WHERE \"On_base_pct\" >= $1 AND \"On_base_pct\" <= $2 AND \"Position\" IN ($3);",
		float64(0), float64(100), "Catcher").Return(nil, nil)
	req := mock.HTTPRequest(t, "GET", "/distil/data", params, query)

	// execute the test request - stubbed ES server will return the JSON
	// loaded above
	res := mock.HTTPResponse(t, req, FilteredDataHandler(storageCtor))
	assert.Equal(t, http.StatusOK, res.Code)

	// compare expected and acutal results - unmarshall first to ensure object
	// rather than byte equality
	expected, err := json.Unmarshal([]byte(
		`{
				"name": "o_185",
				"metadata": [],
				"values": []
			}`))
	assert.NoError(t, err)

	actual, err := json.Unmarshal(res.Body.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func mockContructor(mockDB *mock.DatabaseDriver) pg.ClientCtor {
	return func() (pg.DatabaseDriver, error) {
		return mockDB, nil
	}
}
