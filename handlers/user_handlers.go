package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/services"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/user"
)

type GoogleLogin struct {
	Logger      logger.Logger
	ClientID    string
	RedirectURI string
	Scope       string
}

func (h GoogleLogin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Login Handler")
	defer h.Logger.Info("End Login Handler")

	authURL, err := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
	if err != nil {
		errMsg := fmt.Sprintf("Error - Unable to create auth url. Err: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	sourceURL := r.URL.Query().Get("source_url")

	q := authURL.Query()
	q.Add("scope", h.Scope)
	q.Add("include_granted_scopes", "true")
	q.Add("access_type", "offline")
	q.Add("redirect_uri", h.RedirectURI)
	q.Add("response_type", "code")
	q.Add("client_id", h.ClientID)
	if sourceURL != "" {
		q.Add("state", fmt.Sprintf("source_url=%v", sourceURL))
	}

	authURL.RawQuery = q.Encode()

	http.Redirect(w, r, authURL.String(), http.StatusTemporaryRedirect)
}

type Authenticate struct {
	Logger       logger.Logger
	TableName    string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Auth         services.Auth
	UserStore    user.Store
}

func (h Authenticate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Callback Handler")
	defer h.Logger.Info("End Callback Handler")

	code, ok := r.URL.Query()["code"]
	if !ok {
		errMsg := "Error - Missing code from url param."
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	rawState, ok := r.URL.Query()["state"]
	if !ok {
		h.Logger.Info("No state information is being passed")
	}
	state := ""
	redirectURL := ""
	if ok {
		state = rawState[0]
		h.Logger.Infof("Found transferred state. state: %v", state)
		redirectURL = strings.Split(state, "=")[1]
	}

	h.Logger.Infof("RedirectURL: %v", redirectURL)

	type authRequestBody struct {
		Code         string `json:"code"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RedirectURI  string `json:"redirect_uri"`
		GrantType    string `json:"grant_type"`
	}

	reqBody := authRequestBody{
		Code:         code[0],
		ClientID:     h.ClientID,
		ClientSecret: h.ClientSecret,
		RedirectURI:  h.RedirectURI,
		GrantType:    "authorization_code",
	}

	rawReqBody, _ := json.Marshal(reqBody)

	resp, err := http.Post("https://oauth2.googleapis.com/token", "application/json", bytes.NewBuffer(rawReqBody))
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to receive the input for this request and parse it to json. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	type authResponseBody struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
	}

	rawRespBody, _ := ioutil.ReadAll(resp.Body)
	var authResp authResponseBody
	json.Unmarshal(rawRespBody, &authResp)

	userAPI, _ := url.Parse("https://www.googleapis.com/oauth2/v3/userinfo")
	q := userAPI.Query()
	q.Add("access_token", authResp.AccessToken)
	userAPI.RawQuery = q.Encode()

	userResp, err := http.Get(userAPI.String())
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to receive user information from Google API. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	type userResponseBody struct {
		PictureURL    string `json:"picture"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}

	rawUserRespBody, _ := ioutil.ReadAll(userResp.Body)
	var obtainedUser userResponseBody
	json.Unmarshal(rawUserRespBody, &obtainedUser)

	retrievedUser, err := h.UserStore.GetUserByEmail(context.Background(), obtainedUser.Email)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to obtain user from datastore. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}
	if retrievedUser.ID == "" && retrievedUser.Email == "" {
		id, _ := uuid.NewV4()
		currentTime := time.Now()
		newUser := user.User{
			ID:           id.String(),
			Email:        obtainedUser.Email,
			RefreshToken: authResp.RefreshToken,
			AuthToken:    authResp.AccessToken,
			Type:         "basic",
			DateCreated:  currentTime,
			DateModified: currentTime,
		}
		h.UserStore.Create(context.Background(), newUser)
	}

	token, err := services.NewToken(retrievedUser.ID, h.Auth.ExpiryTime, h.Auth.Secret, h.Auth.Issuer)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to create token. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	type tokenResponse struct {
		Token string `json:"token"`
	}

	rawRespTokenResp, _ := json.Marshal(tokenResponse{Token: token})

	if redirectURL != "" {
		http.Redirect(w, r, redirectURL+fmt.Sprintf("?token=%v", token), http.StatusTemporaryRedirect)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(rawRespTokenResp)
}

// Login - Handles situation of user signing in if user exists - else, throw back error
// Require search of user by email address
type Login struct {
	Logger      logger.Logger
	UserStore   user.Store
	Auth        services.Auth
	RedirectURI string
}

func (h Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Login Handler")
	defer h.Logger.Info("End Login Handler")

	rawReq, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to read json body. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(400)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	type loginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	req := loginReq{}
	err = json.Unmarshal(rawReq, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to parse login body. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(400)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	u, err := h.UserStore.GetUserByEmail(context.TODO(), req.Email)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to find user. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(404)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	passwordCorrect := u.IsPasswordCorrect(req.Password)
	if !passwordCorrect {
		errMsg := fmt.Sprintf("Error - unable to find user. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	// token, err := services.NewToken(u.ID, h.Auth.ExpiryTime, h.Auth.Secret, h.Auth.Issuer)
	// if err != nil {
	// 	errMsg := fmt.Sprintf("Error - unable to create token. Error: %v", err)
	// 	h.Logger.Error(errMsg)
	// 	w.WriteHeader(500)
	// 	w.Write([]byte(generateErrorResp(errMsg)))
	// 	return
	// }

	// type tokenResponse struct {
	// 	Token string `json:"token"`
	// }

	// rawRespTokenResp, _ := json.Marshal(tokenResponse{Token: token})

	value := map[string]string{
		"user_id": u.ID,
	}
	s := securecookie.New(h.Auth.HashKey, h.Auth.BlockKey)
	encoded, err := s.Encode(h.Auth.CookieName, value)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to set authorization token. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}
	cookie := &http.Cookie{
		Name:     h.Auth.CookieName,
		Value:    encoded,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		Expires:  time.Now().Add(1 * time.Hour),
	}
	http.SetCookie(w, cookie)

	// TODO: Handle case of logging in from another page
	if h.RedirectURI != "" {
		http.Redirect(w, r, h.RedirectURI+fmt.Sprintf("?token=%v", ""), http.StatusTemporaryRedirect)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}

// ActivateUser - Handles the sign up url
// Activation link will have 2 query params - user id as well as the user activation token
type ActivateUser struct {
	Logger    logger.Logger
	UserStore user.Store
}

func (h ActivateUser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start ActivateUser Handler")
	defer h.Logger.Info("End ActivateUser Handler")

	token, ok := r.URL.Query()["token"]
	if !ok {
		errMsg := "Error - Invalid activation link"
		h.Logger.Error(errMsg)
		w.WriteHeader(400)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	u, err := h.UserStore.GetUserByActivationToken(context.TODO(), token[0])
	if err != nil {
		errMsg := "Error - Unable to find user"
		h.Logger.Error(errMsg)
		w.WriteHeader(404)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	updateSetters, err := u.Activate(token[0])
	if err != nil {
		errMsg := "Error - Unable to update user's activation"
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	_, err = h.UserStore.Update(context.TODO(), u.ID, updateSetters...)
	if err != nil {
		errMsg := "Error - Unable to update user's activation"
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Activated"))
}

// CreateUser - Handles situation of new sign up to the service
type CreateUser struct {
	Logger    logger.Logger
	UserStore user.Store
}

func (h CreateUser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start CreateUser Handler")
	defer h.Logger.Info("End CreateUser Handler")

	rawReq, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to read json body. Error: %+v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	type createUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	createUserReq := createUserRequest{}
	err = json.Unmarshal(rawReq, &createUserReq)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to parse json body. Error: %+v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	newUser, err := user.New(createUserReq.Email, createUserReq.Password)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to create new user. Error: %+v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	err = h.UserStore.Create(context.TODO(), newUser)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to store newly created user. Error: %+v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	w.WriteHeader(http.StatusCreated)
	type successfulResponse struct {
		Status string `json:"status"`
	}
	rawResp, _ := json.Marshal(successfulResponse{Status: "User Created"})
	w.Write(rawResp)
}

// ForgetPassword issues a link to user via email provider to reset password
type ForgetPassword struct {
	Logger    logger.Logger
	UserStore user.Store
}

func (h ForgetPassword) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start ForgetPassword Handler")
	defer h.Logger.Info("End ForgetPassword Handler")

	rawReq, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to read json body. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(400)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	type forgetPasswordReq struct {
		Email string `json:"email"`
	}
	req := forgetPasswordReq{}
	err = json.Unmarshal(rawReq, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to parse forget password req body. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(400)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	u, err := h.UserStore.GetUserByEmail(context.TODO(), req.Email)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to find user. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(404)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	updateSetters, err := u.ForgetPassword()
	if err != nil {
		errMsg := "Error - Unable to update user's activation"
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	_, err = h.UserStore.Update(context.TODO(), u.ID, updateSetters...)
	if err != nil {
		errMsg := "Error - Unable to update user's forget password token"
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}
}

// ResetPassword - Handles situation where user forget password and needs to reset it
// Reset link will have 2 query params - user id as well as the forget password token
type ResetPassword struct {
	Logger    logger.Logger
	UserStore user.Store
}

func (h ResetPassword) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start ResetPassword Handler")
	defer h.Logger.Info("End ResetPassword Handler")

	rawReq, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to read json body. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(400)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	type resetPasswordReq struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	req := resetPasswordReq{}
	err = json.Unmarshal(rawReq, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to parse login body. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(400)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	u, err := h.UserStore.GetUserByForgetPasswordToken(context.TODO(), req.Token)
	if err != nil {
		errMsg := "Error - Unable to find user"
		h.Logger.Error(errMsg)
		w.WriteHeader(404)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	updateSetters, err := u.ChangePasswordFromForget(req.Token, req.Password)
	if err != nil {
		errMsg := "Error - Unable to update user's change password configuration"
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	_, err = h.UserStore.Update(context.TODO(), u.ID, updateSetters...)
	if err != nil {
		errMsg := "Error - Unable to update user's password"
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}
}

type GetUser struct {
	Logger    logger.Logger
	UserStore user.Store
}

func (h GetUser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start GetUser Handler")
	defer h.Logger.Info("End GetUser Handler")

	userID := mux.Vars(r)["user_id"]
	if userID == "" {
		errMsg := "Missing user id field"
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	u, err := h.UserStore.GetUser(context.TODO(), userID)
	if err != nil {
		errMsg := "Error - Unable to find user"
		h.Logger.Error(errMsg)
		w.WriteHeader(404)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	resp, _ := json.Marshal(u)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
