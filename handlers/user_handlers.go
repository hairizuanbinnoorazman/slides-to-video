package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/gofrs/uuid"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/services"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/user"
)

type Auth struct {
	Secret     string `json:"secret"`
	ExpiryTime int    `json:"expiry_time"`
	Issuer     string `json:"issuer"`
}

type Login struct {
	Logger      logger.Logger
	ClientID    string
	RedirectURI string
	Scope       string
}

func (h Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Login Handler")
	defer h.Logger.Info("End Login Handler")

	authURL, err := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
	if err != nil {
		errMsg := fmt.Sprintf("Error - Unable to create auth url. Err: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	q := authURL.Query()
	q.Add("scope", h.Scope)
	q.Add("include_granted_scopes", "true")
	q.Add("access_type", "offline")
	q.Add("redirect_uri", h.RedirectURI)
	q.Add("response_type", "code")
	q.Add("client_id", h.ClientID)

	authURL.RawQuery = q.Encode()

	http.Redirect(w, r, authURL.String(), http.StatusTemporaryRedirect)
}

type Authenticate struct {
	Logger       logger.Logger
	TableName    string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Auth         Auth
	UserStore    user.Store
}

func (h Authenticate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Callback Handler")
	defer h.Logger.Info("End Callback Handler")

	code, ok := r.URL.Query()["code"]
	if !ok {
		errMsg := fmt.Sprintf("Error - Missing code from url param.")
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

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
		w.Write([]byte(errMsg))
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
		w.Write([]byte(errMsg))
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
		w.Write([]byte(errMsg))
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
		h.UserStore.StoreUser(context.Background(), newUser)
	}

	token, err := services.NewToken(retrievedUser.ID, h.Auth.ExpiryTime, h.Auth.Secret, h.Auth.Issuer)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to create token. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	type tokenResponse struct {
		Token string `json:"token"`
	}

	rawRespTokenResp, _ := json.Marshal(tokenResponse{Token: token})

	w.WriteHeader(http.StatusOK)
	w.Write(rawRespTokenResp)
}
