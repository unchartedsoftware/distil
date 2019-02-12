package datamart

import (
	"strings"

	"github.com/pkg/errors"
	api "github.com/unchartedsoftware/distil/api/model"
)

const (
	// DatasetSuffix is the suffix for the dataset entry when stored in
	// elasticsearch.
	metadataType     = "metadata"
	datasetsListSize = 1000
	// ProvenanceNYU for NYU datamart
	ProvenanceNYU = "datamartNYU"
	// ProvenanceISI for ISI datamart
	ProvenanceISI   = "datamartISI"
	getRESTFunction = "download"
)

// SearchQuery contains the basic properties to query.
type SearchQuery struct {
	Dataset *SearchQueryDatasetProperties `json:"dataset,omitempty"`
}

// SearchQueryDatasetProperties represents queryin on metadata.
type SearchQueryDatasetProperties struct {
	About       string   `json:"about"`
	Description []string `json:"description"`
	Name        []string `json:"name,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
}

// ImportDataset makes the dataset available for ingest and returns
// the URI to use for ingest.
func (s *Storage) ImportDataset(id string, uri string) (string, error) {
	return s.download(s, id, uri)
}

// FetchDatasets returns all datasets in the provided index.
func (s *Storage) FetchDatasets(includeIndex bool, includeMeta bool) ([]*api.Dataset, error) {
	// use default string in search to get complete list
	return s.SearchDatasets("", includeIndex, includeMeta)
}

// FetchDataset returns a dataset in the provided index.
func (s *Storage) FetchDataset(datasetName string, includeIndex bool, includeMeta bool) (*api.Dataset, error) {
	return nil, errors.Errorf("Not implemented")
}

// SearchDatasets returns the datasets that match the search criteria in the
// provided index.
func (s *Storage) SearchDatasets(terms string, includeIndex bool, includeMeta bool) ([]*api.Dataset, error) {
	if terms == "" {
		return make([]*api.Dataset, 0), nil
	}
	return s.searchREST(terms)
}

// SetDataType is not supported by the datamart.
func (s *Storage) SetDataType(dataset string, varName string, varType string) error {
	return errors.Errorf("Not supported")
}

// AddVariable is not supported by the datamart.
func (s *Storage) AddVariable(dataset string, varName string, varType string, varRole string) error {
	return errors.Errorf("Not supported")
}

// DeleteVariable is not supported by the datamart.
func (s *Storage) DeleteVariable(dataset string, varName string) error {
	return errors.Errorf("Not supported")
}

func (s *Storage) searchREST(searchText string) ([]*api.Dataset, error) {
	terms := strings.Fields(searchText)

	// get complete URI for the endpoint
	query := &SearchQuery{
		Dataset: &SearchQueryDatasetProperties{
			About: searchText,
			//Name:        terms,
			Description: terms,
			//Keywords:    terms,
		},
	}

	responseRaw, err := s.search(s, query)
	if err != nil {
		return nil, errors.Wrap(err, "unable to post datamart search request")
	}

	datasets, err := s.parse(responseRaw)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse datamart search response")
	}

	return datasets, nil
}
