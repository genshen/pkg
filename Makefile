PACKAGE=github.com/genshen/pkg/pkg

all: pkg-linux-amd64 pkg-linux-arm64 pkg-darwin-amd64 pkg-windows-amd64.exe

pkg-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pkg-linux-amd64 ${PACKAGE}

pkg-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o pkg-linux-arm64 ${PACKAGE}

pkg-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o pkg-darwin-amd64 ${PACKAGE}

pkg-windows-amd64.exe:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o pkg-windows-amd64.exe ${PACKAGE}

pkg :
	go build -o pkg

clean:
	rm -f pkg-linux-amd64 pkg-linux-arm64 pkg-darwin-amd64 pkg-windows-amd64.exe
