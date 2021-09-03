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

func (server *Server) CreateAdminUser(w http.ResponseWriter, r *http.Request) {

	// get body info
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	// get user model
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// prepares user details for the database insertion
	user.Prepare()

	// validate json fields
	var validations formaterror.GeneralError = user.UserValidations("create", server.DB)
	if len(validations.Errors) > 0 {
		responses.JSON(w, http.StatusUnprocessableEntity, validations)
		return
	}

	userCreated, err := user.SaveUser(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}

	// get user serializer response
	response := userCreated.ShowUserSerializer()

	responses.JSON(w, http.StatusCreated, response)
}

func (server *Server) AddUser(w http.ResponseWriter, r *http.Request) {

	// get body info
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	// get user model
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// prepares user details for the database insertion
	user.Prepare()

	// validate json fields
	var validations formaterror.GeneralError = user.UserValidations("create", server.DB)
	if len(validations.Errors) > 0 {
		responses.JSON(w, http.StatusUnprocessableEntity, validations)
		return
	}

	// get tenant id
	vars := mux.Vars(r)
	tenant_id := vars["tenant_id"]

	// convert tenant id to uuid
	tid_uuid, _ := uuid.Parse(tenant_id)

	userCreated, err := user.SaveUserTenant(server.DB, tid_uuid)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}

	// get user serializer response
	response := userCreated.ShowUserSerializer()

	responses.JSON(w, http.StatusCreated, response)
}

func (server *Server) GetTenantUsers(w http.ResponseWriter, r *http.Request) {

	user := models.User{}

	// get tenant id
	vars := mux.Vars(r)
	tenant_id := vars["tenant_id"]

	users, err := user.FindAllTenantUsers(server.DB, tenant_id, r)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	responses.JSON(w, http.StatusOK, users)
}

func (server *Server) GetTenantUser(w http.ResponseWriter, r *http.Request) {

	user := models.User{}

	// get tenant id
	vars := mux.Vars(r)
	user_id := vars["user_id"]

	userInfo, err := user.FindUserByID(server.DB, user_id)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	responses.JSON(w, http.StatusOK, userInfo.ShowUserSerializer())
}

func (server *Server) ConfirmUser(w http.ResponseWriter, r *http.Request) {

	// get body info
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	// get user model
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// validate json fields
	var validations formaterror.GeneralError = user.UserValidations("confirm", server.DB)
	if len(validations.Errors) > 0 {
		responses.JSON(w, http.StatusUnprocessableEntity, validations)
		return
	}

	users, err := user.ConfirmUser(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, users.ShowUserSerializer())
}
