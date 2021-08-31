package middlewares

import (
	"errors"
	"net/http"

	"siot/api/models"
	"siot/api/responses"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

func SetMiddlewareIsDeviceValid(db *gorm.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// get tenant id
		vars := mux.Vars(r)
		device_id := vars["device_id"]
		tenant_id := vars["tenant_id"]

		// convert tenant and device id to uuid
		tid_uuid, _ := uuid.Parse(tenant_id)
		did_uuid, err := uuid.Parse(device_id)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid device id"))
			return
		}

		device := models.Device{}
		_, err = device.ValidateDevicePermission(db, did_uuid, tid_uuid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("device not found"))
			return
		}

		next(w, r)
	}
}

func SetMiddlewareIsDeviceValidAndActive(db *gorm.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// get device and tenant id
		vars := mux.Vars(r)
		device_id := vars["device_id"]
		tenant_id := vars["tenant_id"]
		secret_key := r.URL.Query().Get("secret_key")

		// convert device and tenant id to uuid
		tid_uuid, _ := uuid.Parse(tenant_id)
		did_uuid, err := uuid.Parse(device_id)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid device id"))
			return
		}

		device := models.Device{}

		// check if device is valid
		hasDevicePerm, err := device.ValidateDevicePermission(db, did_uuid, tid_uuid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("device not found"))
			return
		}

		if secret_key != hasDevicePerm.SecretKey {
			responses.ERROR(w, http.StatusNotFound, errors.New("invalid secret key"))
			return
		}

		// check if device is active
		if hasDevicePerm.Status != "active" {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("device is inactive"))
			return
		}

		next(w, r)
	}
}
