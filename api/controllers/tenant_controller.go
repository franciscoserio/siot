package controllers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"siot/api/auth"
	"siot/api/models"
	"siot/api/responses"
	"siot/api/utils/pagination"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func (server *Server) CreateTenant(w http.ResponseWriter, r *http.Request) {

	// get user token
	user_id, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	// check if user is admin
	// is_admin, _ := server.IsAdmin(user_id)
	// if !is_admin {
	// 	responses.ERROR(w, http.StatusUnauthorized, errors.New("only admin users have access to this endpoint"))
	// 	return
	// }

	// get body info
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	// get tenant model
	tenant := models.Tenant{}
	err = json.Unmarshal(body, &tenant)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// prepares tenant details for the database insertion
	tenant.Prepare()

	// tenant details validation
	validate := validator.New()
	err = validate.Struct(tenant)
	if err != nil {

		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// convert to uuid
	uid_uuid, _ := uuid.Parse(user_id)

	// insert tenant
	tenantCreated, err := tenant.SaveTenant(server.DB, uid_uuid)

	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	responses.JSON(w, http.StatusCreated, tenantCreated)
}

func (server *Server) ListTenants(w http.ResponseWriter, r *http.Request) {

	// get user token
	user_id, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	tenant := models.Tenant{}

	tenants, limit, page, totalRecords, totalPages, nextPage, previousPage, err := tenant.FindAllTenants(server.DB, user_id, r)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	serializerTenants := pagination.ListPaginationSerializer(limit, page, totalRecords, totalPages, nextPage, previousPage, *tenants)
	responses.JSON(w, http.StatusOK, serializerTenants)
}
