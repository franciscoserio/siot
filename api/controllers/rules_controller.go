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

func (server *Server) CreateRule(w http.ResponseWriter, r *http.Request) {

	// get tenant id
	vars := mux.Vars(r)
	tenant_id := vars["tenant_id"]

	// convert tenant id to uuid
	tid_uuid, _ := uuid.Parse(tenant_id)

	// get body info
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	// get rule model
	rule := models.Rule{}
	err = json.Unmarshal(body, &rule)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// validate json fields
	var validations formaterror.GeneralError = rule.RuleValidations(server.DB, tid_uuid)
	if len(validations.Errors) > 0 {
		responses.JSON(w, http.StatusUnprocessableEntity, validations)
		return
	}

	// insert rule
	ruleCreated, err := rule.SaveRule(server.DB, tid_uuid)

	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	responses.JSON(w, http.StatusCreated, ruleCreated)
}

func (server *Server) ListRules(w http.ResponseWriter, r *http.Request) {

	// get tenant id
	vars := mux.Vars(r)
	tenant_id := vars["tenant_id"]

	rule := models.Rule{}

	devices, err := rule.FindAllRules(server.DB, tenant_id, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	responses.JSON(w, http.StatusOK, devices)
}

func (server *Server) ShowRule(w http.ResponseWriter, r *http.Request) {

	// get device id
	vars := mux.Vars(r)
	rule_id := vars["rule_id"]

	rule := models.Rule{}

	ru, err := rule.GetRule(server.DB, rule_id)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, ru)
}

func (server *Server) UpdateRule(w http.ResponseWriter, r *http.Request) {

	// get tenant and rule id
	vars := mux.Vars(r)
	tenant_id := vars["tenant_id"]
	rule_id := vars["rule_id"]

	// convert tenant id to uuid
	tid_uuid, _ := uuid.Parse(tenant_id)

	// get body info
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}

	// get rule model
	rule := models.Rule{}
	err = json.Unmarshal(body, &rule)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// prepares rule details for the database insertion
	rule.PrepareUpdate()

	// validate json fields
	var validations formaterror.GeneralError = rule.RuleValidations(server.DB, tid_uuid)
	if len(validations.Errors) > 0 {
		responses.JSON(w, http.StatusUnprocessableEntity, validations)
		return
	}

	ru, err := rule.UpdateRule(server.DB, rule_id)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, ru)
}

func (server *Server) DeleteRule(w http.ResponseWriter, r *http.Request) {

	// get rule id
	vars := mux.Vars(r)
	rule_id := vars["rule_id"]

	rule := models.Rule{}

	err := rule.DeleteRule(server.DB, rule_id)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
