API_SVC?="slides-to-video-manager"
API_SVC_URL=$$(gcloud beta run services describe $(API_SVC) --region=asia-northeast1 --platform=managed  --format=yaml | yq r - status.address.url)

cloud-run:
	pytest --base-endpoint=$(API_SVC_URL)/api/v1 test_datastore.py test_app.py 