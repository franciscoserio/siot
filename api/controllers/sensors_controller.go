package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"siot/api/models"
	"siot/api/responses"
	"siot/api/utils/formaterror"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (server *Server) CreateSensor(w http.ResponseWriter, r *http.Request) {

	// get body info
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	// get device model
	sensor := models.Sensor{}
	err = json.Unmarshal(body, &sensor)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// get device id
	vars := mux.Vars(r)
	device_id := vars["device_id"]

	// convert tenant id to uuid
	did_uuid, _ := uuid.Parse(device_id)

	// validate json fields
	var validations formaterror.GeneralError = sensor.SensorValidations()
	if len(validations.Errors) > 0 {
		responses.JSON(w, http.StatusUnprocessableEntity, validations)
		return
	}

	// insert sensor
	sensorCreated, err := sensor.SaveSensor(server.DB, did_uuid)

	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	responses.JSON(w, http.StatusCreated, sensorCreated)
}

func (server *Server) ListSensors(w http.ResponseWriter, r *http.Request) {

	// get device id
	vars := mux.Vars(r)
	device_id := vars["device_id"]

	sensor := models.Sensor{}

	devices, err := sensor.FindAllSensors(server.DB, device_id, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	responses.JSON(w, http.StatusOK, devices)
}

func (server *Server) ShowSensor(w http.ResponseWriter, r *http.Request) {

	// get device id
	vars := mux.Vars(r)
	sensor_id := vars["sensor_id"]

	sensor := models.Sensor{}

	d, err := sensor.GetSensor(server.DB, sensor_id)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, d)
}

func (server *Server) UpdateSensor(w http.ResponseWriter, r *http.Request) {

	// get body info
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	// get sensor model
	sensor := models.Sensor{}
	err = json.Unmarshal(body, &sensor)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// prepares device details for the database insertion
	sensor.PrepareUpdate()

	// validate json fields
	var validations formaterror.GeneralError = sensor.SensorValidations()
	if len(validations.Errors) > 0 {
		responses.JSON(w, http.StatusUnprocessableEntity, validations)
		return
	}

	// get device id and sensor id
	vars := mux.Vars(r)
	device_id := vars["device_id"]
	sensor_id := vars["sensor_id"]

	d, err := sensor.UpdateSensor(server.DB, sensor_id, device_id)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, d)
}

func (server *Server) DeleteSensor(w http.ResponseWriter, r *http.Request) {

	// get sensor id
	vars := mux.Vars(r)
	sensor_id := vars["sensor_id"]

	sensor := models.Sensor{}

	err := sensor.DeleteSensor(server.DB, sensor_id)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
