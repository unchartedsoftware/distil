package routes

import (
	"net/http"

	"goji.io/pat"

	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/unchartedsoftware/distil/api/pipeline"
	log "github.com/unchartedsoftware/plog"
)

// PipelineExecuteHandler is a thing that doesn't have the capacity to feel love
func PipelineExecuteHandler(pipelineService *pipeline.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		sessionID := pat.Param(r, "session-id")
		pipelineID := pat.Param(r, "pipeline-id")

		// generate the request
		createReq := pipeline.PipelineExecuteRequest{
			Context:    &pipeline.SessionContext{SessionId: sessionID},
			PipelineId: pipelineID,
		}
		req := pipeline.GeneratePipelineExecuteRequest(&createReq)

		// gets an existing request or dispatchs a new one
		proxy, err := pipelineService.GetOrDispatch(context.Background(), req)
		if err != nil {
			handleError(w, errors.Wrap(err, "failed to issue ExecutePipelineRequest"))
		}

		// process the result proxy, which is replicated for completed, pending requests
		for {
			select {
			case result := <-proxy.Results:
				res := (*result).(*pipeline.PipelineExecuteResult)
				log.Infof("RESULT %v", res.String())
				w.Write([]byte(res.String()))
			case err := <-proxy.Errors:
				log.Info("ERROR")
				handleError(w, errors.Wrap(err, "failed to issue ExecutePipelineRequest"))
			case <-proxy.Done:
				log.Info("DONE")
				return
			}
		}
	}
}
