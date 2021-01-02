APP_VERSION?=$$(git rev-parse --short HEAD)
API_SVC?="slides-to-video-manager"
API_SVC_HOST=$$(gcloud beta run services describe $(API_SVC) --region=asia-northeast1 --platform=managed  --format=yaml | yq r - status.address.url | sed -e "s|https://||")

PDF_SPLITTER_SVC?="pdf-splitter"
PDF_SPLITTER_SVC_URL=$$(gcloud beta run services describe $(PDF_SPLITTER_SVC) --region=asia-northeast1 --platform=managed  --format=yaml | yq r - status.address.url)

IMAGE_TO_VIDEO_SVC?="image-to-video"
IMAGE_TO_VIDEO_SVC_URL=$$(gcloud beta run services describe $(IMAGE_TO_VIDEO_SVC) --region=asia-northeast1 --platform=managed  --format=yaml | yq r - status.address.url)

CONCAT_VIDEO_SVC?="concatenate-video"
CONCAT_VIDEO_SVC_URL=$$(gcloud beta run services describe $(CONCAT_VIDEO_SVC) --region=asia-northeast1 --platform=managed  --format=yaml | yq r - status.address.url)

PROJECT_ID=$$(gcloud config list --format yaml | yq r - core.project)


# Deploy applications to Cloud Run
deploy-mgr:
	gcloud run deploy \
		--image=gcr.io/$(PROJECT_ID)/$(API_SVC):$(APP_VERSION) \
		--service-account=$(API_SVC)@$(PROJECT_ID).iam.gserviceaccount.com \
		--concurrency=20 \
		--cpu=2 \
		--memory=2Gi \
		--platform=managed \
		--allow-unauthenticated \
		--set-env-vars=DATASTORE_GOOGLEDATASTORE_PROJECTID=$(PROJECT_ID),QUEUE_GOOGLEPUBSUB_PROJECTID=$(PROJECT_ID),BLOBSTORAGE_GCS_PROJECTID=$(PROJECT_ID),BLOBSTORAGE_GCS_BUCKET=zontext-pdf-2-videos \
		--region=asia-northeast1 \
		$(API_SVC)

deploy-pdf-splitter:
	gcloud run deploy \
		--image=gcr.io/$(PROJECT_ID)/$(PDF_SPLITTER_SVC):$(APP_VERSION) \
		--service-account=$(PDF_SPLITTER_SVC)@$(PROJECT_ID).iam.gserviceaccount.com \
		--concurrency=1 \
		--cpu=2 \
		--memory=2Gi \
		--platform=managed \
		--no-allow-unauthenticated \
		--region=asia-northeast1 \
		--set-env-vars=SERVER_PORT=8080,BLOBSTORAGE_TYPE=gcs,BLOBSTORAGE_GCS_PROJECTID=$(PROJECT_ID),BLOBSTORAGE_GCS_BUCKET=zontext-pdf-2-videos,SERVER_MANAGERHOST=$(API_SVC_HOST),SERVER_MANAGERPORT=443 \
		--command=/usr/bin/app \
		--args=server \
		$(PDF_SPLITTER_SVC)

deploy-image-to-video:
	gcloud run deploy \
		--image=gcr.io/$(PROJECT_ID)/$(IMAGE_TO_VIDEO_SVC):$(APP_VERSION) \
		--service-account=$(IMAGE_TO_VIDEO_SVC)@$(PROJECT_ID).iam.gserviceaccount.com \
		--concurrency=1 \
		--cpu=2 \
		--memory=2Gi \
		--no-allow-unauthenticated \
		--platform=managed \
		--region=asia-northeast1

deploy-concatenate-video:
	gcloud run deploy \
		--image=gcr.io/$(PROJECT_ID)/$(CONCAT_VIDEO_SVC):$(APP_VERSION) \
		--service-account=$(CONCAT_VIDEO_SVC)@$(PROJECT_ID).iam.gserviceaccount.com \
		--concurrency=1 \
		--cpu=2 \
		--memory=2Gi \
		--platform=managed \
		--region=asia-northeast1

link-pdf-splitter:
	gcloud run services add-iam-policy-binding $(PDF_SPLITTER_SVC) \
   		--member=serviceAccount:$(PDF_SPLITTER_SVC)@${PROJECT_ID}.iam.gserviceaccount.com \
		--region=asia-northeast1 \
		--platform=managed \
   		--role=roles/run.invoker
	gcloud pubsub topics create $(PDF_SPLITTER_SVC)
	gcloud pubsub subscriptions create $(PDF_SPLITTER_SVC) --topic $(PDF_SPLITTER_SVC) \
		--ack-deadline 300 \
		--push-endpoint=$(PDF_SPLITTER_SVC_URL) \
		--push-auth-service-account=$(PDF_SPLITTER_SVC)@${PROJECT_ID}.iam.gserviceaccount.com

link-image-to-video:
	gcloud beta run services add-iam-policy-binding $(IMAGE_TO_VIDEO_SVC) \
   		--member=serviceAccount:cloud-run-pubsub-invoker@${PROJECT_ID}.iam.gserviceaccount.com \
		--region=asia-northeast1 \
		--platform=managed \
   		--role=roles/run.invoker
	gcloud pubsub topics create $(IMAGE_TO_VIDEO_SVC)
	gcloud beta pubsub subscriptions create $(IMAGE_TO_VIDEO_SVC) --topic $(IMAGE_TO_VIDEO_SVC) \
		--ack-deadline 300 \
		--push-endpoint=$(IMAGE_TO_VIDEO_SVC_URL) \
		--push-auth-service-account=cloud-run-pubsub-invoker@${PROJECT_ID}.iam.gserviceaccount.com

link-concatenate-video:
	gcloud beta run services add-iam-policy-binding $(CONCAT_VIDEO_SVC) \
   		--member=serviceAccount:cloud-run-pubsub-invoker@${PROJECT_ID}.iam.gserviceaccount.com \
		--region=asia-northeast1 \
		--platform=managed \
   		--role=roles/run.invoker
	gcloud pubsub topics create $(CONCAT_VIDEO_SVC)
	gcloud beta pubsub subscriptions create $(CONCAT_VIDEO_SVC) --topic $(CONCAT_VIDEO_SVC) \
		--ack-deadline 300 \
		--push-endpoint=$(CONCAT_VIDEO_SVC_URL) \
		--push-auth-service-account=cloud-run-pubsub-invoker@${PROJECT_ID}.iam.gserviceaccount.com

drop-services:
	gcloud pubsub subscriptions delete $(PDF_SPLITTER_SVC)
	gcloud pubsub topics delete $(PDF_SPLITTER_SVC)
	gcloud run services delete $(API_SVC) --platform=managed --region=asia-northeast1 --quiet
	gcloud run services delete $(PDF_SPLITTER_SVC) --platform=managed --region=asia-northeast1 --quiet
	