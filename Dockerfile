## pkg
FROM golang:1.10.1

MAINTAINER genshen genshenchu@gmail.com

ARG PROJECT_PATH="/go/src/github.com/genshen/pkg"
ARG BINARY_NAME="pkg"

# Workdir is path in your docker image from where all your commands will be executed
WORKDIR ${PROJECT_PATH}

# Add all from your project inside workdir of docker image
ADD . ${PROJECT_PATH}

# Then run your script to install dependencies and build application
RUN go get -u github.com/kardianos/govendor \
    && govendor sync \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ${BINARY_NAME}

# Next start another building context
FROM scratch

# Copy only build result from previous step to new lightweight image
COPY --from=0 ${PROJECT_PATH}/${BINARY_NAME} .

### copy pkg file.
# COPY pkg /usr/local/bin

# CMD ["/bin/ash"]