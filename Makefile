PACKAGE=github.com/genshen/pkg/pkg

GIT_COMMIT=$(shell git rev-parse HEAD)
BUILD_TIME=$(shell date "+%F %T")

define LD_FLAGS
-ldflags "-X 'github.com/genshen/pkg/pkg/version.GitCommitID=${GIT_COMMIT}' -X 'github.com/genshen/pkg/pkg/version.BuildTime=${BUILD_TIME}'"
endef

all: pkg-linux-amd64 pkg-linux-arm64 pkg-darwin-amd64 pkg-darwin-arm64 pkg-windows-amd64.exe

pkg-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LD_FLAGS} -o pkg-linux-amd64 ${PACKAGE}

pkg-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ${LD_FLAGS} -o pkg-linux-arm64 ${PACKAGE}

pkg-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ${LD_FLAGS} -o pkg-darwin-amd64 ${PACKAGE}

pkg-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build ${LD_FLAGS} -o pkg-darwin-arm64 ${PACKAGE}

pkg-windows-amd64.exe:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ${LD_FLAGS} -o pkg-windows-amd64.exe ${PACKAGE}

pkg :
	go build -o pkg ${PACKAGE}

clean:
	rm -f pkg-linux-amd64 pkg-linux-arm64 pkg-darwin-amd64 pkg-darwin-arm64 pkg-windows-amd64.exe
