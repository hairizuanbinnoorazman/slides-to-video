FROM ubuntu:20.04 as prod
# Added ca-certificates as that is needed for golang libs to contact google api servers. W/o it, it would suffer from ssl issues
RUN apt update && apt install -y imagemagick ca-certificates
# Ghostscript issues - imagemagick has been set such that it would disable features based on a policy file. PDF manipulation is one of them
RUN sed -i 's/<policy domain="coder" rights="none"/<policy domain="coder" rights="read"/g' /etc/ImageMagick-6/policy.xml
COPY app /usr/bin/app
WORKDIR /usr/bin
EXPOSE 8080
CMD ["app"]