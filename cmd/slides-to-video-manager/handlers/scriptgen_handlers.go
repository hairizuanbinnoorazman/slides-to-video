package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/acl"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/blobstorage"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/logger"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/project"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/scriptgenerator"
	"github.com/hairizuanbinnoorazman/slides-to-video-manager/videosegment"
)

type GenerateProjectScripts struct {
	Logger            logger.Logger
	ProjectStore      project.Store
	VideoSegmentStore videosegment.Store
	ACLStore          acl.Store
	BlobStorage       blobstorage.BlobStorage
	ScriptGenerator   scriptgenerator.ScriptGenerator
	ImagesFolder      string
}

func (h GenerateProjectScripts) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start GenerateProjectScripts API Handler")
	defer h.Logger.Info("End GenerateProjectScripts API Handler")

	ctx := r.Context()
	projectID := mux.Vars(r)["project_id"]
	userID := ctx.Value(userIDKey).(string)

	if h.ScriptGenerator == nil {
		errMsg := "Script generation is not configured on this server"
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	obtainedACL, err := h.ACLStore.Get(ctx, projectID, userID)
	if err != nil || !obtainedACL.IsAuthorized(acl.Editor) {
		errMsg := fmt.Sprintf("Error - unable to confirm acl for project. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	singleProject, err := h.ProjectStore.Get(ctx, projectID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve project. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	segments := singleProject.VideoSegments
	sort.Sort(videosegment.ByOrder(segments))

	// If project has no description, generate one from all slide images
	if singleProject.Description == "" {
		var allSlides []scriptgenerator.SlideImage
		for _, seg := range segments {
			if seg.ImageID == "" {
				continue
			}
			imagePath := fmt.Sprintf("%s/%s", h.ImagesFolder, seg.ImageID)
			content, loadErr := h.BlobStorage.Load(ctx, imagePath)
			if loadErr != nil {
				h.Logger.Errorf("Error loading image for description generation. ImageID: %s, Error: %v", seg.ImageID, loadErr)
				continue
			}
			allSlides = append(allSlides, scriptgenerator.SlideImage{
				ImageID: seg.ImageID,
				Order:   seg.Order,
				Content: content,
			})
		}

		if len(allSlides) > 0 {
			desc, genErr := h.ScriptGenerator.GenerateDescription(ctx, allSlides)
			if genErr != nil {
				errMsg := fmt.Sprintf("Error - unable to generate project description. Error: %v", genErr)
				h.Logger.Error(errMsg)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(generateErrorResp(errMsg)))
				return
			}

			descSetters, _ := project.SetDescription(desc)
			singleProject, err = h.ProjectStore.Update(ctx, projectID, descSetters...)
			if err != nil {
				errMsg := fmt.Sprintf("Error - unable to save generated description. Error: %v", err)
				h.Logger.Error(errMsg)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(generateErrorResp(errMsg)))
				return
			}
		}
	}

	slidesUpdated := 0
	slidesSkipped := 0

	for _, seg := range segments {
		if seg.Script != "" {
			slidesSkipped++
			continue
		}
		if seg.ImageID == "" {
			slidesSkipped++
			continue
		}

		imagePath := fmt.Sprintf("%s/%s", h.ImagesFolder, seg.ImageID)
		content, loadErr := h.BlobStorage.Load(ctx, imagePath)
		if loadErr != nil {
			h.Logger.Errorf("Error loading image for script generation. ImageID: %s, Error: %v", seg.ImageID, loadErr)
			slidesSkipped++
			continue
		}

		slide := scriptgenerator.SlideImage{
			ImageID: seg.ImageID,
			Order:   seg.Order,
			Content: content,
		}

		script, genErr := h.ScriptGenerator.GenerateScript(ctx, singleProject.Description, slide)
		if genErr != nil {
			h.Logger.Errorf("Error generating script for slide. ImageID: %s, Error: %v", seg.ImageID, genErr)
			slidesSkipped++
			continue
		}

		scriptSetters, _ := videosegment.SetScript(script)
		_, updateErr := h.VideoSegmentStore.Update(ctx, projectID, seg.ID, scriptSetters...)
		if updateErr != nil {
			h.Logger.Errorf("Error saving generated script. VideoSegmentID: %s, Error: %v", seg.ID, updateErr)
			slidesSkipped++
			continue
		}

		slidesUpdated++
	}

	type generateScriptsResp struct {
		SlidesUpdated int    `json:"slides_updated"`
		SlidesSkipped int    `json:"slides_skipped"`
		Description   string `json:"description"`
	}

	resp := generateScriptsResp{
		SlidesUpdated: slidesUpdated,
		SlidesSkipped: slidesSkipped,
		Description:   singleProject.Description,
	}
	rawResp, _ := json.Marshal(resp)

	w.WriteHeader(http.StatusOK)
	w.Write(rawResp)
}

type GenerateVideoSegmentScript struct {
	Logger            logger.Logger
	ProjectStore      project.Store
	VideoSegmentStore videosegment.Store
	ACLStore          acl.Store
	BlobStorage       blobstorage.BlobStorage
	ScriptGenerator   scriptgenerator.ScriptGenerator
	ImagesFolder      string
}

func (h GenerateVideoSegmentScript) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start GenerateVideoSegmentScript API Handler")
	defer h.Logger.Info("End GenerateVideoSegmentScript API Handler")

	ctx := r.Context()
	projectID := mux.Vars(r)["project_id"]
	videoSegmentID := mux.Vars(r)["videosegment_id"]
	userID := ctx.Value(userIDKey).(string)

	if h.ScriptGenerator == nil {
		errMsg := "Script generation is not configured on this server"
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	obtainedACL, err := h.ACLStore.Get(ctx, projectID, userID)
	if err != nil || !obtainedACL.IsAuthorized(acl.Editor) {
		errMsg := fmt.Sprintf("Error - unable to confirm acl for project. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	singleProject, err := h.ProjectStore.Get(ctx, projectID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve project. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	if singleProject.Description == "" {
		errMsg := "Project description is empty. Set a description first or use bulk generation (POST :generate-scripts)"
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	seg, err := h.VideoSegmentStore.Get(ctx, projectID, videoSegmentID)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to retrieve video segment. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	if seg.ImageID == "" {
		errMsg := "Video segment has no associated image"
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	imagePath := fmt.Sprintf("%s/%s", h.ImagesFolder, seg.ImageID)
	content, err := h.BlobStorage.Load(ctx, imagePath)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to load slide image. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	slide := scriptgenerator.SlideImage{
		ImageID: seg.ImageID,
		Order:   seg.Order,
		Content: content,
	}

	script, err := h.ScriptGenerator.GenerateScript(ctx, singleProject.Description, slide)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to generate script. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	scriptSetters, _ := videosegment.SetScript(script)
	updatedSeg, err := h.VideoSegmentStore.Update(ctx, projectID, videoSegmentID, scriptSetters...)
	if err != nil {
		errMsg := fmt.Sprintf("Error - unable to save generated script. Error: %v", err)
		h.Logger.Error(errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(generateErrorResp(errMsg)))
		return
	}

	rawResp, _ := json.Marshal(updatedSeg)
	w.WriteHeader(http.StatusOK)
	w.Write(rawResp)
}
