package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"siot/api/models"
	"siot/api/responses"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (server *Server) SendData(w http.ResponseWriter, r *http.Request) {

	// get body info
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	// get data model
	data := models.Data{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// validate
	validate := validator.New()
	err = validate.Struct(data)
	if err != nil {

		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// get device id
	vars := mux.Vars(r)
	device_id := vars["device_id"]

	// convert device id to uuid
	did_uuid, _ := uuid.Parse(device_id)

	// prepares data for insertion
	err_validation := data.ValidateAndSendData(server.MDB, server.DB, did_uuid)
	if err_validation != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err_validation)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (server *Server) GetData(w http.ResponseWriter, r *http.Request) {

	// get device id
	vars := mux.Vars(r)
	device_id := vars["device_id"]

	// convert device id to uuid
	did_uuid, _ := uuid.Parse(device_id)

	data := models.Data{}
	result, err := data.GetData(server.MDB, server.DB, did_uuid, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	if result.Data == nil {
		result.Data = make([]map[string]interface{}, 0)
	}

	responses.JSON(w, http.StatusOK, result)
}
