//
//   Copyright © 2021 Uncharted Software Inc.
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

package compute

import (
	"context"
	"encoding/json"

	"github.com/uncharted-distil/distil-compute/primitive/compute"
)

// StopSolutionSearchRequest represents a request to stop any pending siolution searches.
type StopSolutionSearchRequest struct {
	RequestID string `json:"requestId"`
}

// NewStopSolutionSearchRequest instantiates a new StopSolutionSearchRequest.
func NewStopSolutionSearchRequest(data []byte) (*StopSolutionSearchRequest, error) {
	req := &StopSolutionSearchRequest{}
	err := json.Unmarshal(data, &req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Dispatch dispatches the stop search request.
func (s *StopSolutionSearchRequest) Dispatch(client *compute.Client) error {
	return client.StopSearch(context.Background(), s.RequestID)
}
