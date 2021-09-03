package controllers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"siot/api/auth"
	"siot/api/models"
	"siot/api/responses"
	"siot/api/utils/formaterror"
	"siot/api/utils/pagination"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (server *Server) CreateTenant(w http.ResponseWriter, r *http.Request) {

	// get user token
	user_id, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

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

	// validate json fields
	var validations formaterror.GeneralError = tenant.TenantValidations()
	if len(validations.Errors) > 0 {
		responses.JSON(w, http.StatusUnprocessableEntity, validations)
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

func (server *Server) GetTenant(w http.ResponseWriter, r *http.Request) {

	// get tenant id
	vars := mux.Vars(r)
	tenant_id := vars["tenant_id"]

	tenant := models.Tenant{}

	t, err := tenant.GetTenant(server.DB, tenant_id)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	responses.JSON(w, http.StatusOK, t)
}

func (server *Server) UpdateTenant(w http.ResponseWriter, r *http.Request) {

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

	// get tenant id
	vars := mux.Vars(r)
	tenant_id := vars["tenant_id"]

	// prepares device details for the database insertion
	tenant.PrepareUpdate()

	t, err := tenant.UpdateTenant(server.DB, tenant_id)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	responses.JSON(w, http.StatusOK, t)
}
