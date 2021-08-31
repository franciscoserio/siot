package middlewares

import (
	"errors"
	"net/http"

	"siot/api/auth"
	"siot/api/responses"
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

func SetMiddlewareIsAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		is_valid := auth.ExtractIsAdmin(r)
		if !is_valid {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("only admin users have access to this endpoint"))
			return
		}

		next(w, r)
	}
}
