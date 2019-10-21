FROM golang:1.13 as builder
WORKDIR /go/src/github.com/hairizuanbinnoorazman/slides-to-video-manager
COPY go.mod .
COPY go.sum .
RUN go mod download
ADD . .
RUN go build -o app .

FROM ubuntu:18.10 as prod
RUN apt update && apt install -y ca-certificates
COPY --from=builder /go/src/github.com/hairizuanbinnoorazman/slides-to-video-manager/app /usr/bin/app
ADD slides-to-video-manager.json /usr/bin/slides-to-video-manager.json
ADD main.html /usr/bin/main.html
ADD parentJobs.html /usr/bin/parentJobs.html
WORKDIR /usr/bin
EXPOSE 8080
CMD ["app"]
