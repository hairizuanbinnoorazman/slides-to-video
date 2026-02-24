FROM ubuntu:20.04 as prod

# Install all dependencies needed for all-in-one mode (embedded workers)
# - ca-certificates: Required for Go TLS connections to Google APIs
# - imagemagick: Required by pdf-splitter worker to extract slides from PDFs
# - ffmpeg: Required by image-to-video and concatenate-video workers
RUN apt update && apt install -y ca-certificates imagemagick ffmpeg

# Fix ImageMagick policy to allow PDF operations
# ImageMagick disables PDF operations by default due to Ghostscript vulnerabilities
# Change from rights="none" to rights="read" for coder domain
RUN sed -i 's/<policy domain="coder" rights="none"/<policy domain="coder" rights="read"/g' /etc/ImageMagick-6/policy.xml

COPY app /usr/bin/app
WORKDIR /usr/bin
EXPOSE 8080
CMD ["app", "server"]
