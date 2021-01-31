FROM ubuntu:20.04 as prod
RUN apt update && apt install -y ca-certificates
COPY app /usr/bin/app
WORKDIR /usr/bin
EXPOSE 8080
CMD ["app", "server"]
