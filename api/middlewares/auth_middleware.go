package middlewares

import (
	"errors"
	"net/http"

	"siot/api/auth"
	"siot/api/models"
	"siot/api/responses"

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
