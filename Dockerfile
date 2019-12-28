## pkg with clang toolchain and cmake (pkg may use cc/cxx compiler and cmake tool).

# docker build --rm=true  -t genshen/pkg:0.2.0 .

FROM golang:1.13.5-alpine3.11 AS builder

MAINTAINER genshen genshenchu@gmail.com

ARG PROJECT_PATH="/go/src/github.com/genshen/pkg"
ARG PACKAFE_NAME="github.com/genshen/pkg/pkg"
ARG BINARY_NAME="pkg"

RUN apk add --no-cache git

# Add all from your project inside workdir of docker image
ADD . ${PROJECT_PATH}

# Then run your script to install dependencies and build application
RUN cd ${PROJECT_PATH} \
    && go mod download \
    && CGO_ENABLED=0 GOOS=linux go build -o ${GOPATH}/bin/${BINARY_NAME} ${PACKAFE_NAME}


## build cmake from source
FROM genshen/clang-toolchain:9.0.0 AS cmake_builder

ARG OPENSSL_DOOWNLOOAD_URL="https://cdn.openbsd.org/pub/OpenBSD/LibreSSL/libressl-3.0.2.tar.gz"

# we need remove cmake help
ARG CMAKE_DOWNLOAD_URL="https://cmake.org/files/v3.16/cmake-3.16.2.tar.gz"
ARG CMAKE_INSATLL_PATH=/usr/local/cmake
ARG CMAKE_HELP_PATH=share/cmake-3.16/Help

# build libressl
# wget is already install in alpine
RUN apk add --no-cache make wget \
    && wget ${OPENSSL_DOOWNLOOAD_URL} -O /tmp/libressl.tar.gz \
    && mkdir -p /tmp/libressl-src \
    && tar zxf /tmp/libressl.tar.gz -C /tmp/libressl-src  --strip-components=1 \
    && cd /tmp/libressl-src \
    && ./configure --prefix=/usr/local/libressl CC=clang \
    && make -j$(nproc) \
    && make install \
    && rm -rf /tmp/libressl-src /tmp/libressl.tar.gz \
    # remove static lib files
    && rm -rf /usr/local/libressl/lib/*.a /usr/local/libressl/lib/*.la \
    && ln -s /usr/local/libressl/bin/* /usr/local/bin/ \
    && ln -s /usr/local/libressl/lib/* /usr/local/lib/ \
    && ln -s /usr/local/libressl/include/* /usr/local/include/

RUN wget ${CMAKE_DOWNLOAD_URL} -O /tmp/cmake.tar.gz \
    && mkdir -p /tmp/cmake-src \
    && tar zxf /tmp/cmake.tar.gz -C /tmp/cmake-src  --strip-components=1 \
    && cd /tmp/cmake-src \
    && ./bootstrap --parallel=$(nproc) --prefix=${CMAKE_INSATLL_PATH} -- -DCMAKE_BUILD_TYPE:STRING=Release \
    && make -j$(nproc) \
    && make install \
    && rm -rf ${CMAKE_INSATLL_PATH}/${CMAKE_HELP_PATH} \
    && rm -rf /tmp/cmake-src /tmp/cmake.tar.gz

# build ninja
#todo add ninja


# next start another building context
FROM genshen/clang-toolchain:9.0.0
# Copy only build result from previous step to new lightweight image
COPY --from=cmake_builder /usr/local/cmake /usr/local/cmake
COPY --from=cmake_builder /usr/local/libressl/lib/ /usr/local/libressl/
COPY --from=builder /go/bin/pkg /usr/local/bin/pkg

RUN mkdir -p /usr/local/bin /usr/local/lib \
    && ln -s /usr/local/cmake/bin/* /usr/local/bin/ \
    && ln -s /usr/local/libressl/* /usr/local/lib/ \
    && apk add --no-cache make

ENTRYPOINT ["/usr/local/bin/pkg"]
CMD ["--help"]
