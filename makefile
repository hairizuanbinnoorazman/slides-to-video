####
# Prepare to migrate commands to deployment repo
####

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

build-images: 
	docker build -t slides-to-video-manager ./cmd/slides-to-video-manager
	docker build -t pdf-splitter ./cmd/pdf-splitter
	docker build -t image-to-video ./cmd/image-to-video

build-all: build-bin build-images

stack-up:
	cd docker-compose && docker-compose up

stack-down:
	cd docker-compose && docker-compose down

stack-up-monitoring:
	cd docker-compose && docker-compose -f docker-compose.yaml -f with-monitoring.yaml up

stack-down-monitoring:
	cd docker-compose && docker-compose -f docker-compose.yaml -f with-monitoring.yaml down
