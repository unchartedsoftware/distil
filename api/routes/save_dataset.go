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

package routes

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/task"
	"goji.io/v3/pat"
)

// SaveDatasetHandler extracts a dataset from storage and writes it to disk.
func SaveDatasetHandler(metaCtor api.MetadataStorageCtor, dataCtor api.DataStorageCtor, config env.Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// get dataset name
		dataset := pat.Param(r, "dataset")
		invert := parseBoolParam(pat.Param(r, "invert"))
		params, err := getPostParameters(r)
		if err != nil {
			handleError(w, errors.Wrap(err, "Unable to parse post parameters"))
			return
		}
		// get variable names and ranges out of the params
		filterParams, err := api.ParseFilterParamsFromJSON(params)
		if err != nil {
			handleError(w, err)
			return
		}
		// get storage clients
		metaStorage, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		dataStorage, err := dataCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		// replace any grouped variables in filter params with the group's
		expandedFilterParams, err := api.ExpandFilterParams(dataset, filterParams, false, metaStorage)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable to expand filter params"))
			return
		}
		ds, err := metaStorage.FetchDataset(dataset, true, true, true)
		if err != nil {
			handleError(w, err)
			return
		}
		if ds.Immutable{
			handleError(w, errors.New("can not mutate an immutable dataset"))
			return
		}
		_, _, err = task.ExportDataset(dataset, metaStorage, dataStorage, invert, expandedFilterParams)
		if err != nil {
			handleError(w, err)
			return
		}
		// delete rows based on filterParams
		err = dataStorage.SaveDataset(dataset, ds.StorageName, invert, expandedFilterParams)
		if err != nil {
			handleError(w, err)
			return
		}
		// version dataset
		ds.Immutable = true
		// is no longer a clone due to the dropping of the filterParams
		ds.Clone = false
		err = metaStorage.UpdateDataset(ds)
		if err != nil {
			handleError(w, err)
			return
		}
	}
}
