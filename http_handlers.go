package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hairizuanbinnoorazman/slides-to-video-manager/services"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/user"

	"github.com/gofrs/uuid"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
)

type scriptParse struct {
	Script []string `json:"script"`
}

type exampleHandler struct {
	logger           Logger
	client           *storage.Client
	datastoreClient  *datastore.Client
	pubsubClient     *pubsub.Client
	bucketName       string
	bucketFolderName string
	parentTableName  string
	tableName        string
	topicName        string
}

func (h exampleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Example Handler")
	defer h.logger.Info("End Example Handler")

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve parse multipart form data. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}
	script := r.FormValue("script")
	h.logger.Info(script)
	file, handler, err := r.FormFile("myfile")
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve form data. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}
	defer file.Close()
	var b bytes.Buffer
	bw := bufio.NewWriter(&b)
	io.Copy(bw, file)

	bs := BlobStorage{
		logger:     h.logger,
		client:     h.client,
		bucketName: h.bucketName,
	}
	parentJob, err := h.CreateParentJob(handler.Filename, script)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to save parent job to datastore. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	bs.Save(context.Background(), h.bucketFolderName+"/"+parentJob.Filename, b.Bytes())

	job, err := h.CreatePDFSplitJob(parentJob.ID, parentJob.Filename)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to save pdf split job. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	values := map[string]string{"id": job.ID, "pdfFileName": job.Filename}
	jsonValue, _ := json.Marshal(values)
	pubsub := Pubsub{h.logger, h.pubsubClient, h.topicName}
	err = pubsub.publish(context.Background(), jsonValue)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to send pdf split job. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%v successfully uploaded. Go to the /jobs page for viewing progress", handler.Filename)))
	return
}

func (h exampleHandler) CreateParentJob(filename, script string) (ParentJob, error) {
	store := NewStore(h.datastoreClient, h.parentTableName)
	rawID, _ := uuid.NewV4()
	jobID := rawID.String()
	parentJob := ParentJob{
		ID:               jobID,
		OriginalFilename: filename,
		Filename:         jobID + ".pdf",
		Script:           script,
		Status:           "created",
	}
	err := store.StoreParentJob(context.Background(), parentJob)
	if err != nil {
		return ParentJob{}, err
	}
	return parentJob, nil
}

func (h exampleHandler) CreatePDFSplitJob(parentJobID, filename string) (PDFToImageJob, error) {
	store := NewStore(h.datastoreClient, h.tableName)
	rawID, _ := uuid.NewV4()
	jobID := rawID.String()
	pdfToImageJob := PDFToImageJob{
		ID:          jobID,
		ParentJobID: parentJobID,
		Filename:    filename,
		Status:      "created",
	}
	err := store.StorePDFToImageJob(context.Background(), pdfToImageJob)
	if err != nil {
		return PDFToImageJob{}, err
	}
	return pdfToImageJob, nil
}

type mainPage struct {
	logger Logger
}

func (h mainPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Main Page Handler")
	defer h.logger.Info("End Main Page Handler")

	t, err := template.ParseFiles("main.html")
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to parse templete. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
	}

	err = t.Execute(w, nil)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to parse templete. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
	}
	return
}

type reportPDFSplit struct {
	logger          Logger
	datastoreClient *datastore.Client
	pubsubClient    *pubsub.Client
	parentTableName string
	tableName       string
	nextTableName   string
	nextTopicName   string
}

type SlideDetail struct {
	ImageID string `json:"image"`
	SlideNo int    `json:"slide_no"`
}

func (h reportPDFSplit) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Report PDF Split Handler")
	defer h.logger.Info("End Report PDF Split Handler")

	type request struct {
		ID           string        `json:"id"`
		SlideDetails []SlideDetail `json:"slide_details"`
		Status       string        `json:"status"`
	}

	var req request
	rawData, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to receive the input for this request and parse it to json. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
	}

	store := NewStore(h.datastoreClient, h.tableName)
	job, err := store.GetPDFToImageJob(context.Background(), req.ID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to get pdf to image job details. Error: %v %v", err, h.tableName)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	if job.Status == "completed" {
		h.logger.Infof("Detected completed PDF job but another request wants to reset it to running")
		w.WriteHeader(200)
		w.Write([]byte("Completed"))
		return
	}

	job.Status = req.Status
	err = store.StorePDFToImageJob(context.Background(), job)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to save and store pdf status. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	parentStore := NewStore(h.datastoreClient, h.parentTableName)
	parentJob, _ := parentStore.GetParentJob(context.Background(), job.ParentJobID)
	var scripts scriptParse
	json.Unmarshal([]byte(parentJob.Script), &scripts)

	parentJob.Status = "running"
	parentStore.StoreParentJob(context.Background(), parentJob)

	for i, slideDetail := range req.SlideDetails {
		store := NewStore(h.datastoreClient, h.nextTableName)
		rawID, _ := uuid.NewV4()
		jobID := rawID.String()
		splitFileName := strings.Split(slideDetail.ImageID, "-")
		slideNoAndFileFormat := strings.Split(splitFileName[len(splitFileName)-1], ".")
		num, _ := strconv.Atoi(slideNoAndFileFormat[0])

		image2videoJob := ImageToVideoJob{
			ID:          jobID,
			ParentJobID: job.ParentJobID,
			ImageID:     slideDetail.ImageID,
			SlideID:     num,
			Text:        scripts.Script[i],
			Status:      "created",
		}
		store.StoreImageToVideoJob(context.Background(), image2videoJob)

		values := map[string]string{"id": image2videoJob.ID, "image_id": image2videoJob.ImageID, "text": image2videoJob.Text}
		jsonValue, _ := json.Marshal(values)
		pubsub := Pubsub{h.logger, h.pubsubClient, h.nextTopicName}
		err = pubsub.publish(context.Background(), jsonValue)
		if err != nil {
			errMsg := fmt.Sprintf("Error - unable to send pdf split job. Error: %v", err)
			h.logger.Error(errMsg)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Implemented"))
}

type reportImageToVideo struct {
	logger          Logger
	datastoreClient *datastore.Client
	pubsubClient    *pubsub.Client
	tableName       string
	nextTableName   string
	nextTopicName   string
}

func (h reportImageToVideo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Report Image To Video Handler")
	defer h.logger.Info("End Report Image to Video Handler")

	type request struct {
		ID         string `json:"id"`
		Status     string `json:"status"`
		OutputFile string `json:"output_file"`
	}

	var req request
	rawData, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to receive the input for this request and parse it to json. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	store := NewStore(h.datastoreClient, h.tableName)
	job, err := store.GetImageToVideoJob(context.Background(), req.ID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to get pdf to image to video details. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	if job.Status == "completed" {
		h.logger.Infof("Detected completed Image to Video job but another request wants to reset it to running")
		w.WriteHeader(200)
		w.Write([]byte("Completed"))
		return
	}

	job.Status = req.Status
	job.OutputFile = req.OutputFile
	store.StoreImageToVideoJob(context.Background(), job)

	items, err := store.GetAllImageToVideoJobs(context.Background(), job.ParentJobID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve list of job ids based on parent. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	completedJobs := 0
	videoList := []string{}
	for _, item := range items {
		if item.Status == "completed" {
			completedJobs = completedJobs + 1
			videoList = append(videoList, item.OutputFile)
		}
	}
	if len(items) != completedJobs {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Not all jobs completed, will not issue new one"))
		return
	}

	videoList = []string{}
	for i, _ := range items {
		for _, item := range items {
			if item.SlideID == i {
				videoList = append(videoList, item.OutputFile)
			}
		}
	}

	nextStore := NewStore(h.datastoreClient, h.nextTableName)
	rawID, _ := uuid.NewV4()
	concatJob := VideoConcatJob{
		ID:          rawID.String(),
		ParentJobID: job.ParentJobID,
		Videos:      videoList,
		Status:      "created",
	}
	nextStore.StoreVideoConcatJob(context.Background(), concatJob)

	values := map[string]interface{}{"id": concatJob.ID, "video_ids": concatJob.Videos}
	jsonValue, _ := json.Marshal(values)
	pubsub := Pubsub{h.logger, h.pubsubClient, h.nextTopicName}
	err = pubsub.publish(context.Background(), jsonValue)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to send pdf split job. Error: %v", err)
		h.logger.Error(errMsg)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Stored New value"))
}

type reportVideoConcat struct {
	logger          Logger
	datastoreClient *datastore.Client
	tableName       string
	parentTableName string
}

func (h reportVideoConcat) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Report Video Concat Handler")
	defer h.logger.Info("End Report Video Concat Handler")

	type request struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		OutputVideo string `json:"output_video"`
	}

	var req request
	rawData, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to receive the input for this request and parse it to json. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	store := NewStore(h.datastoreClient, h.tableName)
	job, err := store.GetVideoConcatJob(context.Background(), req.ID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to get pdf to image job details. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	parentStore := NewStore(h.datastoreClient, h.parentTableName)
	parentJob, _ := parentStore.GetParentJob(context.Background(), job.ParentJobID)

	if parentJob.Status == "completed" {
		h.logger.Infof("Detected parent job status is already completed but another request comes in to reset it. Reject it")
		w.WriteHeader(200)
		w.Write([]byte("Completed"))
		return
	}

	job.Status = req.Status
	job.OutputFile = req.OutputVideo

	store.StoreVideoConcatJob(context.Background(), job)

	parentJob.Status = req.Status
	parentJob.VideoFile = req.OutputVideo

	parentStore.StoreParentJob(context.Background(), parentJob)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Implemented"))
}

type viewAllParentJobs struct {
	logger          Logger
	datastoreClient *datastore.Client
	tableName       string
}

func (h viewAllParentJobs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start View All Parent Jobs Handler")
	defer h.logger.Info("End View All Parent Jobs Handler")

	store := NewStore(h.datastoreClient, h.tableName)
	parentJobs, err := store.GetAllParentJobs(context.Background())
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to view all parent jobs. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	t, err := template.ParseFiles("parentJobs.html")
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to parse templete. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	varmap := map[string]interface{}{
		"parentJobs": parentJobs,
	}

	err = t.Execute(w, varmap)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to parse templete. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}
	return
}

type viewAllParentJobsAPI struct {
	logger          Logger
	datastoreClient *datastore.Client
	tableName       string
}

func (h viewAllParentJobsAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start View All Parent Jobs API Handler")
	defer h.logger.Info("End View All Parent Jobs API Handler")

	store := NewStore(h.datastoreClient, h.tableName)
	parentJobs, err := store.GetAllParentJobs(context.Background())
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to view all parent jobs. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	type jobsResponse struct {
		Jobs []ParentJob `json:"jobs"`
	}

	rawParentJobs, err := json.Marshal(jobsResponse{Jobs: parentJobs})
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to view all parent jobs. Error: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(rawParentJobs)
	return
}

type downloadJob struct {
	logger        Logger
	storageClient *storage.Client
	bucketName    string
}

func (h downloadJob) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Download Handler")
	defer h.logger.Info("End Download Handler")

	filename, ok := r.URL.Query()["filename"]
	if !ok {
		errMsg := fmt.Sprintf("Error - Missing filename from url param.")
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	bs := BlobStorage{h.logger, h.storageClient, h.bucketName}
	content, err := bs.Load(context.Background(), "videos/"+filename[0])
	if err != nil {
		errMsg := fmt.Sprintf("Error - Unable to download file from blob storage. Err: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(200)
	w.Write(content)
}

type login struct {
	logger      Logger
	clientID    string
	redirectURI string
	scope       string
}

func (h login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Login Handler")
	defer h.logger.Info("End Login Handler")

	authURL, err := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
	if err != nil {
		errMsg := fmt.Sprintf("Error - Unable to create auth url. Err: %v", err)
		h.logger.Error(errMsg)
		w.WriteHeader(500)
		w.Write([]byte(errMsg))
		return
	}

	q := authURL.Query()
	q.Add("scope", h.scope)
	q.Add("include_granted_scopes", "true")
	q.Add("access_type", "offline")
	q.Add("redirect_uri", h.redirectURI)
	q.Add("response_type", "code")
	q.Add("client_id", h.clientID)

	authURL.RawQuery = q.Encode()

	http.Redirect(w, r, authURL.String(), http.StatusTemporaryRedirect)
}

type authenticate struct {
	logger       Logger
	tableName    string
	clientID     string
	clientSecret string
	redirectURI  string
	auth         Auth
	userStore    user.UserStore
}

func (h authenticate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Start Callback Handler")
	defer h.logger.Info("End Callback Handler")

	code, ok := r.URL.Query()["code"]
	if !ok {
		errMsg := fmt.Sprintf("Error - Missing code from url param.")
		h.logger.Error(errMsg)
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
		ClientID:     h.clientID,
		ClientSecret: h.clientSecret,
		RedirectURI:  h.redirectURI,
		GrantType:    "authorization_code",
	}

	rawReqBody, _ := json.Marshal(reqBody)

	resp, err := http.Post("https://oauth2.googleapis.com/token", "application/json", bytes.NewBuffer(rawReqBody))
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to receive the input for this request and parse it to json. Error: %v", err)
		h.logger.Error(errMsg)
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
		h.logger.Error(errMsg)
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

	retrievedUser, err := h.userStore.GetUserByEmail(context.Background(), obtainedUser.Email)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to obtain user from datastore. Error: %v", err)
		h.logger.Error(errMsg)
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
		h.userStore.StoreUser(context.Background(), newUser)
	}

	token, err := services.NewToken(retrievedUser.ID, h.auth.ExpiryTime, h.auth.Secret, h.auth.Issuer)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to create token. Error: %v", err)
		h.logger.Error(errMsg)
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
