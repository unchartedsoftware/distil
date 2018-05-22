package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/plog"

	"github.com/unchartedsoftware/distil/api/compute"
	"github.com/unchartedsoftware/distil/api/model"
	jutil "github.com/unchartedsoftware/distil/api/util/json"
)

const (
	createSolutions   = "CREATE_SOLUTIONS"
	stopSolutions     = "STOP_SOLUTIONS"
	categoricalType   = "categorical"
	numericalType     = "numerical"
	defaultResourceID = "0"
	datasetSizeLimit  = 10000
)

// SolutionHandler represents a solution websocket handler.
func SolutionHandler(client *compute.Client, metadataCtor model.MetadataStorageCtor, dataCtor model.DataStorageCtor, solutionCtor model.SolutionStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// create conn
		conn, err := NewConnection(w, r, handleSolutionMessage(client, metadataCtor, dataCtor, solutionCtor))
		if err != nil {
			log.Warn(err)
			return
		}
		// listen for requests and respond
		err = conn.ListenAndRespond()
		if err != nil {
			log.Info(err)
		}
		// clean up conn internals
		conn.Close()
	}
}

func handleSolutionMessage(client *compute.Client, metadataCtor model.MetadataStorageCtor, dataCtor model.DataStorageCtor, solutionCtor model.SolutionStorageCtor) func(conn *Connection, bytes []byte) {
	return func(conn *Connection, bytes []byte) {
		// parse the message
		msg, err := NewMessage(bytes)
		if err != nil {
			// parsing error, send back a failure response
			err := fmt.Errorf("unable to parse solution request message: %s", string(bytes))
			// send error response
			handleErr(conn, nil, err)
			return
		}
		// handle message
		go handleMessage(conn, client, metadataCtor, dataCtor, solutionCtor, msg)
	}
}

func parseMessage(bytes []byte) (*Message, error) {
	var msg *Message
	err := json.Unmarshal(bytes, &msg)
	if err != nil {
		return nil, err
	}
	msg.Timestamp = time.Now()
	return msg, nil
}

func handleMessage(conn *Connection, client *compute.Client, metadataCtor model.MetadataStorageCtor, dataCtor model.DataStorageCtor, solutionCtor model.SolutionStorageCtor, msg *Message) {
	switch msg.Type {
	case createSolutions:
		handleCreateSolutions(conn, client, metadataCtor, dataCtor, solutionCtor, msg)
		return
	case stopSolutions:
		handleStopSolutions(conn, client, msg)
		return
	default:
		// unrecognized type
		handleErr(conn, msg, errors.New("unrecognized message type"))
		return
	}
}

func handleCreateSolutions(conn *Connection, client *compute.Client, metadataCtor model.MetadataStorageCtor, dataCtor model.DataStorageCtor, solutionCtor model.SolutionStorageCtor, msg *Message) {
	// unmarshal request
	request, err := compute.NewSolutionRequest(msg.Raw)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// initialize the storage
	dataStorage, err := dataCtor()
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// initialize metadata storage
	metaStorage, err := metadataCtor()
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// initialize solution storage
	solutionStorage, err := solutionCtor()
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// persist the request information and dispatch the request
	err = request.PersistAndDispatch(client, solutionStorage, metaStorage, dataStorage)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// listen for solution updates
	err = request.Listen(func(status compute.SolutionStatus) {
		// check for error
		if status.Error != nil {
			handleErr(conn, msg, err)
			return
		}
		// send status to client
		handleSuccess(conn, msg, jutil.StructToMap(status))
	})
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// complete the request
	handleComplete(conn, msg)
}

func handleStopSolutions(conn *Connection, client *compute.Client, msg *Message) {
	// unmarshal request
	request, err := compute.NewStopSolutionSearchRequest(msg.Raw)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}

	// dispatch request
	err = request.Dispatch(client)
	if err != nil {
		handleErr(conn, msg, err)
		return
	}
	// complete the request
	handleComplete(conn, msg)
	return
}
