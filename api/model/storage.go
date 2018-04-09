package model

import (
	"time"
)

// DataStorageCtor represents a client constructor to instantiate a data
// storage client.
type DataStorageCtor func() (DataStorage, error)

// DataStorage defines the functions available to query the underlying data storage.
type DataStorage interface {
	FetchNumRows(dataset string, filters map[string]interface{}) (int, error)
	FetchData(dataset string, index string, filterParams *FilterParams, invert bool) (*FilteredData, error)
	FetchSummary(dataset string, index string, varName string, filterParams *FilterParams) (*Histogram, error)
	FetchSummaryByResult(dataset string, index string, varName string, resultURI string, filterParams *FilterParams, extrema *Extrema) (*Histogram, error)
	PersistResult(dataset string, resultURI string) error
	FetchResults(dataset string, index string, resultURI string) (*FilteredData, error)
	FetchFilteredResults(dataset string, index string, resultURI string, filterParams *FilterParams) (*FilteredData, error)
	FetchResultsSummary(dataset string, resultURI string, index string, filterParams *FilterParams, extrema *Extrema) (*Histogram, error)
	FetchResultsExtremaByURI(dataset string, resultURI string, index string) (*Extrema, error)
	FetchResidualsSummary(dataset string, resultURI string, index string, filterParams *FilterParams, extrema *Extrema) (*Histogram, error)
	FetchResidualsExtremaByURI(dataset string, resultURI string, index string) (*Extrema, error)
	FetchExtremaByURI(dataset string, resultURI string, index string, variable string) (*Extrema, error)

	// Dataset manipulation
	SetDataType(dataset string, index string, field string, fieldType string) error
}

// PipelineStorageCtor represents a client constructor to instantiate a
// pipeline storage client.
type PipelineStorageCtor func() (PipelineStorage, error)

// PipelineStorage defines the functions available to query the underlying
// pipeline storage.
type PipelineStorage interface {
	PersistModel(modelID string, dataset string, progress string, createdTime time.Time) error
	PersistModelFeature(modelID string, featureName string, featureType string) error
	PersistModelFilters(modelID string, filters *FilterParams) error
	PersistPipeline(modelID string, pipelineID string, progress string, createdTime time.Time) error
	PersistPipelineResult(pipelineID string, resultUUID string, resultURI string, progress string, createdTime time.Time) error
	PersistPipelineScore(pipelineID string, metric string, score float64) error
	UpdateModel(modelID string, progress string, updatedTime time.Time) error
	FetchModel(modelID string) (*Model, error)
	FetchModelFeatures(modelID string) ([]*Feature, error)
	FetchModelFilters(modelID string, features []*Feature) (*FilterParams, error)
	FetchPipeline(pipelineID string) (*Pipeline, error)
	FetchPipelineResultByModelID(modelID string) ([]*PipelineResult, error)
	FetchPipelineResultByUUID(resultUUID string) (*PipelineResult, error)
	FetchPipelineResult(pipelineID string) (*PipelineResult, error)
	FetchPipelineResultByDatasetTarget(dataset string, target string, pipelineID string) ([]*PipelineResult, error)
	FetchPipelineScore(pipelineID string) ([]*PipelineScore, error)
}


// MetadataStorageCtor represents a client constructor to instantiate a
// metadata storage client.
type MetadataStorageCtor func() (MetadataStorage, error)

// MetadataStorage defines the functions available to query the underlying
// metadata storage.
type MetadataStorage interface {
	FetchVariables(dataset string, index string, includeIndex bool) ([]*Variable, error)
	FetchVariablesDisplay(dataset string, index string) ([]*Variable, error)
	FetchVariable(dataset string, index string, varName string) (*Variable, error)
	FetchVariableDisplay(dataset string, index string, varName string) (*Variable, error)
	FetchDatasets(index string, includeIndex bool) ([]*Dataset, error)
	SearchDatasets(index string, terms string, includeIndex bool) ([]*Dataset, error)

	// Dataset manipulation
	SetDataType(dataset string, index string, field string, fieldType string) error
}
