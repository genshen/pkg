## pkg
# docker build --rm=true  -t genshen/pkg:0.2.0 .
FROM golang:1.12 as builder

MAINTAINER genshen genshenchu@gmail.com

ARG PROJECT_PATH="/go/src/github.com/genshen/pkg"
ARG PACKAFE_NAME="github.com/genshen/pkg/cmds"
ARG BINARY_NAME="pkg"

# Add all from your project inside workdir of docker image
ADD . ${PROJECT_PATH}

# Then run your script to install dependencies and build application
RUN CGO_ENABLED=0 GOOS=linux go build -o ${BINARY_NAME} ${PACKAFE_NAME}

# Next start another building context
FROM alpine:latest

# Copy only build result from previous step to new lightweight image
COPY --from=builder /go/src/github.com/genshen/pkg/pkg /usr/local/bin/pkg
#ADD --chown=root:root https://github.com/genshen/pkg/releases/download/v0.2.0-beta/pkg-linux-amd64 /usr/local/bin/pkg

RUN sudo chmod 755 /usr/local/bin/pkg & sudo apk add --no-cache cmake

CMD ["/bin/ash"]
