package middlewares

import (
	"errors"
	"net/http"

	"siot/api/auth"
	"siot/api/models"
	"siot/api/responses"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

func SetMiddlewareIsTenantValid(db *gorm.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		// get user token
		user_id, err := auth.ExtractTokenID(r)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
			return
		}

		// convert user id to uuid
		uid_uuid, _ := uuid.Parse(user_id)

		// get tenant id
		vars := mux.Vars(r)
		tenant_id := vars["tenant_id"]

		// convert tenant id to uuid
		tid_uuid, err := uuid.Parse(tenant_id)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid tenant id"))
			return
		}

		user_tenant := models.UserTenant{}
		hasTenantPerm := user_tenant.ValidateTenantPermission(db, uid_uuid, tid_uuid)
		if hasTenantPerm == -1 {
			responses.ERROR(w, http.StatusNotFound, errors.New("tenant not found"))
			return

		} else if hasTenantPerm == -2 {
			responses.ERROR(w, http.StatusNotFound, errors.New("you do not have permissions on this tenant"))
			return

		} else if hasTenantPerm == -3 {
			responses.ERROR(w, http.StatusNotFound, errors.New("tenant is inactive"))
			return

		}

		next(w, r)
	}
}
