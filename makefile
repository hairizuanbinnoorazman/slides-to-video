####
# Prepare to migrate commands to deployment repo
####

git_hash=$$(git rev-parse --short HEAD)
image_repo?=""
image_version?=latest

setup:
	@echo "Run setup steps"
	~/go/bin/goimports main.go > main_temp.go && mv main_temp.go main.go
	~/go/bin/goimports http_handlers.go > http_handlers_temp.go && mv http_handlers_temp.go http_handlers.go
	
encrypt:
	gcloud kms encrypt --ciphertext-file=slides-to-video-manager.json.enc --plaintext-file=slides-to-video-manager.json --location=global --keyring=test --key=test1
	gcloud kms encrypt --ciphertext-file=config.json.enc --plaintext-file=config.json --location=global --keyring=test --key=test1

build-bin: 
	GOOS=linux GOARCH=amd64 go build -o ./cmd/slides-to-video-manager/app ./cmd/slides-to-video-manager
	GOOS=linux GOARCH=amd64 go build -o ./cmd/pdf-splitter/app ./cmd/pdf-splitter
	GOOS=linux GOARCH=amd64 go build -o ./cmd/image-to-video/app ./cmd/image-to-video
	GOOS=linux GOARCH=amd64 go build -o ./cmd/concatenate-video/app ./cmd/concatenate-video
	GOOS=linux GOARCH=amd64 go build -o ./cmd/slides-to-video-frontend/app ./cmd/slides-to-video-frontend

build-images: 
	docker build -t $(image_repo)slides-to-video-manager:$(image_version) ./cmd/slides-to-video-manager
	docker build -t $(image_repo)pdf-splitter:$(image_version) ./cmd/pdf-splitter
	docker build -t $(image_repo)image-to-video:$(image_version) ./cmd/image-to-video
	docker build -t $(image_repo)concatenate-video:$(image_version) ./cmd/concatenate-video
	docker build -t $(image_repo)slides-to-video-frontend:$(image_version) ./cmd/slides-to-video-frontend

push-images:
	docker push $(image_repo)slides-to-video-manager:$(image_version)
	docker push $(image_repo)pdf-splitter:$(image_version)
	docker push $(image_repo)image-to-video:$(image_version)
	docker push $(image_repo)concatenate-video:$(image_version)
	docker push $(image_repo)slides-to-video-frontend:$(image_version)

build-all: build-bin build-images

build-all-versioned:
	$(eval image_version := $(shell git rev-parse --short HEAD))
	$(eval image_repo := gcr.io/$(shell gcloud config list --format yaml | yq r - core.project))
	make image_version=$(image_version) image_repo=$(image_repo) build-all

push-all-versioned:
	$(eval image_version := $(shell git rev-parse --short HEAD))
	$(eval image_repo := gcr.io/$(shell gcloud config list --format yaml | yq r - core.project))
	make image_version=$(image_version) image_repo=$(image_repo) push-images

stack-up:
	cd deployment/docker-compose && docker-compose up

stack-down:
	cd deployment/docker-compose && docker-compose down

stack-up-monitoring:
	cd deployment/docker-compose && docker-compose -f docker-compose.yaml -f with-monitoring.yaml up

stack-down-monitoring:
	cd deployment/docker-compose && docker-compose -f docker-compose.yaml -f with-monitoring.yaml down
