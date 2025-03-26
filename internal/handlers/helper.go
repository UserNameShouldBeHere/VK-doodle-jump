package handlers

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Errors string `json:"errors"`
}

type ResponseData struct {
	Status int
	Data   any
}

func WriteResponse(w http.ResponseWriter, responseData ResponseData) error {
	jsonData, err := json.Marshal(responseData)
	if err != nil {
		return err
	}

	w.WriteHeader(responseData.Status)
	_, err = w.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}
