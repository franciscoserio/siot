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

func (server *Server) CreateDevice(w http.ResponseWriter, r *http.Request) {

	// get tenant id
	vars := mux.Vars(r)
	tenant_id := vars["tenant_id"]

	// convert tenant id to uuid
	tid_uuid, _ := uuid.Parse(tenant_id)

	// get body info
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	// get device model
	device := models.Device{}
	err = json.Unmarshal(body, &device)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// validate json fields
	var validations formaterror.GeneralError = device.DeviceValidations()
	if len(validations.Errors) > 0 {
		responses.JSON(w, http.StatusUnprocessableEntity, validations)
		return
	}

	// insert device
	deviceCreated, err := device.SaveDevice(server.MDB, server.DB, tid_uuid)

	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	serializedDevice, _ := device.FindDevice(server.DB, deviceCreated.ID)

	responses.JSON(w, http.StatusCreated, serializedDevice)
}

func (server *Server) ListDevices(w http.ResponseWriter, r *http.Request) {

	// get tenant id
	vars := mux.Vars(r)
	tenant_id := vars["tenant_id"]

	device := models.Device{}

	devices, err := device.FindAllDevices(server.DB, tenant_id, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	responses.JSON(w, http.StatusOK, devices)
}

func (server *Server) ShowDevice(w http.ResponseWriter, r *http.Request) {

	// get device id
	vars := mux.Vars(r)
	device_id := vars["device_id"]

	// convert tenant id to uuid
	did_uuid, _ := uuid.Parse(device_id)

	device := models.Device{}

	d, err := device.FindDevice(server.DB, did_uuid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, d)
}

func (server *Server) UpdateDevice(w http.ResponseWriter, r *http.Request) {

	// get body info
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	// get device model
	device := models.Device{}
	err = json.Unmarshal(body, &device)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// prepares device details for the database insertion
	device.PrepareUpdate()

	// get device id
	vars := mux.Vars(r)
	device_id := vars["device_id"]

	// convert tenant id to uuid
	did_uuid, _ := uuid.Parse(device_id)

	d, err := device.UpdateDevice(server.DB, did_uuid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, d)
}

func (server *Server) DeleteDevice(w http.ResponseWriter, r *http.Request) {

	// get device id
	vars := mux.Vars(r)
	device_id := vars["device_id"]

	// convert tenant id to uuid
	did_uuid, _ := uuid.Parse(device_id)

	device := models.Device{}

	err := device.DeleteDevice(server.DB, did_uuid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
