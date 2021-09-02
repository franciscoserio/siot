package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"siot/api/auth"
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

func (server *Server) GetUser(w http.ResponseWriter, r *http.Request) {

	// vars := mux.Vars(r)
	// uid, err := vars["id"]
	// if err != nil {
	// 	responses.ERROR(w, http.StatusBadRequest, err)
	// 	return
	// }
	// user := models.User{}
	// userGotten, err := user.FindUserByID(server.DB, uid)
	// if err != nil {
	// 	responses.ERROR(w, http.StatusBadRequest, err)
	// 	return
	// }
	// responses.JSON(w, http.StatusOK, userGotten)
	responses.JSON(w, http.StatusOK, nil)
}

// func (server *Server) UpdateUser(w http.ResponseWriter, r *http.Request) {

// 	vars := mux.Vars(r)
// 	uid, err := strconv.ParseUint(vars["id"], 10, 32)
// 	if err != nil {
// 		responses.ERROR(w, http.StatusBadRequest, err)
// 		return
// 	}
// 	body, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		responses.ERROR(w, http.StatusUnprocessableEntity, err)
// 		return
// 	}
// 	user := models.User{}
// 	err = json.Unmarshal(body, &user)
// 	if err != nil {
// 		responses.ERROR(w, http.StatusUnprocessableEntity, err)
// 		return
// 	}
// 	tokenID, err := auth.ExtractTokenID(r)
// 	if err != nil {
// 		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
// 		return
// 	}
// 	if tokenID != fmt.Sprint(uid) {
// 		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
// 		return
// 	}
// 	user.Prepare()
// 	var errors []string = user.Validate("login")
// 	if len(errors) > 0 {
// 		responses.ERROR(w, http.StatusUnprocessableEntity, err)
// 		return
// 	}
// 	updatedUser, err := user.UpdateAUser(server.DB, uint32(uid))
// 	if err != nil {
// 		formattedError := formaterror.FormatError(err.Error())
// 		responses.ERROR(w, http.StatusInternalServerError, formattedError)
// 		return
// 	}
// 	responses.JSON(w, http.StatusOK, updatedUser)
// }

func (server *Server) DeleteUser(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	user := models.User{}

	uid, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	tokenID, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	if tokenID != "" && tokenID != fmt.Sprint(uid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New(http.StatusText(http.StatusUnauthorized)))
		return
	}
	_, err = user.DeleteAUser(server.DB, uint32(uid))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", uid))
	responses.JSON(w, http.StatusNoContent, "")
}
