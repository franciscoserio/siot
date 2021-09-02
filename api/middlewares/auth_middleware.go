package middlewares

import (
	"errors"
	"net/http"

	"siot/api/auth"
	"siot/api/models"
	"siot/api/responses"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

func SetMiddlewareJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next(w, r)
	}
}

func SetMiddlewareAuthentication(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		err := auth.TokenValid(r)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid token"))
			return
		}
		next(w, r)
	}
}

func SetMiddlewareIsSuperAdmin(db *gorm.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user_id, err := auth.ExtractTokenID(r)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
			return
		}

		user := models.User{}
		u, err := user.FindUserByID(db, user_id)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
			return
		}

		if !u.IsSuperAdmin {
			responses.ERROR(w, http.StatusNotFound, errors.New("only super admin users can access this endpoint"))
			return
		}

		next(w, r)
	}
}

func SetMiddlewareIsAdmin(db *gorm.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user_id, err := auth.ExtractTokenID(r)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
			return
		}

		user := models.User{}
		u, err := user.FindUserByID(db, user_id)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
			return
		}

		if !u.IsAdmin {
			responses.ERROR(w, http.StatusNotFound, errors.New("only super admin users can access this endpoint"))
			return
		}

		next(w, r)
	}
}

func SetMiddlewareIsUserTenantValid(db *gorm.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// get tenant id
		vars := mux.Vars(r)
		user_id := vars["user_id"]
		tenant_id := vars["tenant_id"]

		user := models.User{}
		u := user.BelongsToTenant(db, tenant_id, user_id)

		if !u {
			responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
			return
		}

		next(w, r)
	}
}
