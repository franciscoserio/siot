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

func SetMiddlewareIsSensorValid(db *gorm.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// get tenant id
		vars := mux.Vars(r)
		device_id := vars["device_id"]
		sensor_id := vars["sensor_id"]

		// convert device and sensor id to uuid
		did_uuid, _ := uuid.Parse(device_id)
		sid_uuid, err := uuid.Parse(sensor_id)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid sensor id"))
			return
		}

		sensor := models.Sensor{}

		isSensorValid, _ := sensor.IsValidSensor(db, sid_uuid, did_uuid)

		if !isSensorValid {
			responses.ERROR(w, http.StatusNotFound, errors.New("sensor not found"))
			return
		}

		next(w, r)
	}
}
