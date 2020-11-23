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

build-mgr:
	GOOS=linux GOARCH=amd64 go build -o ./cmd/slides-to-video-manager/app ./cmd/slides-to-video-manager

build-bin: build-mgr

build-image-mgr:
	docker build -t slides-to-video-manager ./cmd/slides-to-video-manager

build-images: build-image-mgr

build-all: build-bin build-images
