package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"siot/api/auth"
	"siot/api/models"
	"siot/api/responses"
	"siot/api/serializers"
	"siot/api/utils/formaterror"
)

func (server *Server) Login(w http.ResponseWriter, r *http.Request) {

	// get json body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// validate json fields
	var validationErrors []string = user.Validate("login")
	if len(validationErrors) > 0 {
		var strErrors []string
		for i := range validationErrors {
			strErrors = append(strErrors, validationErrors[i])
		}
		errorResponse := map[string]interface{}{
			"errors": strErrors,
		}
		responses.JSON(w, http.StatusBadRequest, errorResponse)

		return
	}

	// get token
	token, err := server.SignIn(user.Email, user.Password)
	if err != nil {
		formattedError := formaterror.LoginError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}

	// user information
	userDetails, err := user.FindUserByEmail(server.DB, user.Email)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	var resp serializers.LoginSerializer
	resp.Token = token
	resp.User = userDetails.ShowUserSerializer()

	responses.JSON(w, http.StatusOK, resp)
}

func (server *Server) SignIn(email, password string) (string, error) {

	var err error

	user := models.User{}

	err = server.DB.Model(models.User{}).Where("email = ?", email).Take(&user).Error
	if err != nil {
		return "", err
	}
	err = models.VerifyPassword(user.Password, password)
	if err != nil {
		return "", err
	}
	return auth.CreateToken(user.ID, user.IsAdmin, user.Status)
}
