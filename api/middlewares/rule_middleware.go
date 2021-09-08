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

func SetMiddlewareIsRuleValid(db *gorm.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// get tenant id
		vars := mux.Vars(r)
		tenant_id := vars["tenant_id"]
		rule_id := vars["rule_id"]

		// convert tenant and sensor id to uuid
		tid_uuid, _ := uuid.Parse(tenant_id)
		rid_uuid, err := uuid.Parse(rule_id)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid rule id"))
			return
		}

		rule := models.Rule{}

		isRuleValid, _ := rule.IsValidRule(db, tid_uuid, rid_uuid)

		if !isRuleValid {
			responses.ERROR(w, http.StatusNotFound, errors.New("rule not found"))
			return
		}

		next(w, r)
	}
}
