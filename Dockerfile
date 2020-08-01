FROM golang:1.13 as builder
WORKDIR /go/src/github.com/hairizuanbinnoorazman/slides-to-video-manager
COPY go.mod .
COPY go.sum .
RUN go mod download
ADD . .
RUN go build -o app ./cmd/slides-to-video-manager/

FROM ubuntu:20.04 as prod
RUN apt update && apt install -y ca-certificates
COPY --from=builder /go/src/github.com/hairizuanbinnoorazman/slides-to-video-manager/app /usr/bin/app
WORKDIR /usr/bin
EXPOSE 8080
CMD ["app"]
