package server

import (
	"encoding/json"
	"github.com/apibrew/rapid-api/pkg/data"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *Server) setupRestApi(r *mux.Router) {
	r.Methods("GET").HandlerFunc(s.getHandler)
	r.Methods("POST").HandlerFunc(s.writeHandler)
	r.Methods("DELETE").HandlerFunc(s.deleteHandler1)
}

func (s *Server) getHandler(writer http.ResponseWriter, request *http.Request) {
	var path = request.URL.Path

	records, isCollection, err := s.DataInterface.GetRecords(path)

	if err != nil {
		handleServerError(writer, err)
		return
	}

	if len(records) == 0 {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	writer.Header().Set("Content-Type", "application/json")

	if isCollection {
		err = json.NewEncoder(writer).Encode(records)
		if err != nil {
			handleServerError(writer, err)
			return
		}
	} else {
		err = json.NewEncoder(writer).Encode(records[0])
		if err != nil {
			handleServerError(writer, err)
			return
		}
	}
}

func (s *Server) deleteHandler1(writer http.ResponseWriter, request *http.Request) {
	var path = request.URL.Path

	err := s.DataInterface.DeleteRecord(path)

	if err != nil {
		handleServerError(writer, err)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (s *Server) writeHandler(writer http.ResponseWriter, request *http.Request) {
	var path = request.URL.Path

	var record data.Record

	err := json.NewDecoder(request.Body).Decode(&record)

	if err != nil {
		handleClientError(writer, err)
		return
	}

	delete(record, "path")

	record, err = s.DataInterface.WriteRecord(path, record)

	if err != nil {
		handleServerError(writer, err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(writer).Encode(record)

	if err != nil {
		handleServerError(writer, err)
		return
	}
}

func handleServerError(writer http.ResponseWriter, err error) {
	writer.WriteHeader(http.StatusInternalServerError)
	_, _ = writer.Write([]byte(err.Error()))
}

func handleClientError(writer http.ResponseWriter, err error) {
	writer.WriteHeader(http.StatusBadRequest)
	_, _ = writer.Write([]byte(err.Error()))
}
