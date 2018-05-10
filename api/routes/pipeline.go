package routes

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"goji.io/pat"

	"github.com/unchartedsoftware/distil/api/model"
)

// SolutionInfo represents the solution information relevant to the client.
type SolutionInfo struct {
	RequestID   string                 `json:"requestId"`
	Feature     string                 `json:"feature"`
	SolutionID  string                 `json:"solutionId"`
	ResultUUID  string                 `json:"resultId"`
	Progress    string                 `json:"progress"`
	Scores      []*model.SolutionScore `json:"scores"`
	CreatedTime time.Time              `json:"timestamp"`
	Dataset     string                 `json:"dataset"`
	Features    []*model.Feature       `json:"features"`
	Filters     *model.FilterParams    `json:"filters"`
}

// SolutionResponse represents a request response
type SolutionResponse struct {
	Solutions []*SolutionInfo `json:"solutions"`
}

// SolutionHandler fetches existing solutions.
func SolutionHandler(solutionCtor model.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// extract route parameters
		dataset := pat.Param(r, "dataset")
		target := pat.Param(r, "target")
		solutionID := pat.Param(r, "solution-id")

		if solutionID == "null" {
			solutionID = ""
		}
		if dataset == "null" {
			dataset = ""
		}
		if target == "null" {
			target = ""
		}

		solution, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		requests, err := solution.FetchSolutionResultByDatasetTarget(dataset, target, solutionID)
		if err != nil {
			handleError(w, err)
			return
		}

		// flatten the results
		solutions := make([]*SolutionInfo, 0)
		for _, req := range requests {

			for _, pip := range req.Solutions {
				solution := &SolutionInfo{
					// request
					RequestID: req.RequestID,
					Dataset:   req.Dataset,
					Feature:   req.TargetFeature(),
					Features:  req.Features,
					Filters:   req.Filters,
					// solution
					SolutionID:  pip.SolutionID,
					Scores:      pip.Scores,
					CreatedTime: pip.CreatedTime,
					Progress:    pip.Progress,
				}
				for _, res := range pip.Results {
					// result
					solution.CreatedTime = res.CreatedTime
					solution.ResultUUID = res.ResultUUID
					solution.Progress = res.Progress
				}

				solutions = append(solutions, solution)
			}
		}

		// marshall data and sent the response back
		err = handleJSON(w, &SolutionResponse{
			Solutions: solutions,
		})
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal session solutions into JSON"))
			return
		}

		return
	}
}
