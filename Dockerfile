# create image from the official Go image
FROM golang

# Create binary directory, install glide and fresh
RUN go get github.com/pilu/fresh

# define work directory
ADD . /go/src/gowiki
WORKDIR /go/src/gowiki

EXPOSE 8080

# serve the app
CMD fresh -c runner.conf
