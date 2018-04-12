## pkg
# docker build --rm=true  -t genshen/pkg:0.0.1 .
FROM golang:1.10.1 as builder

MAINTAINER genshen genshenchu@gmail.com

ARG PROJECT_PATH="/go/src/github.com/genshen/pkg"
ARG BINARY_NAME="pkg"

# Workdir is path in your docker image from where all your commands will be executed
WORKDIR ${PROJECT_PATH}

# Add all from your project inside workdir of docker image
ADD . ${PROJECT_PATH}

# Then run your script to install dependencies and build application
RUN go get -u -v github.com/kardianos/govendor \
    && govendor sync \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ${BINARY_NAME} .

# Next start another building context
FROM alpine:latest

# Copy only build result from previous step to new lightweight image
COPY --from=builder /go/src/github.com/genshen/pkg/pkg /usr/local/bin/pkg

CMD ["/bin/ash"]
