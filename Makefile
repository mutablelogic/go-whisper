# Paths to packages
DOCKER=$(shell which docker)
GIT=$(shell which git)

# Set OS and Architecture
ARCH ?= $(shell arch | tr A-Z a-z | sed 's/x86_64/amd64/' | sed 's/i386/amd64/' | sed 's/armv7l/arm/' | sed 's/aarch64/arm64/')
OS ?= $(shell uname | tr A-Z a-z)
VERSION ?= $(shell git describe --tags --always | sed 's/^v//')
DOCKER_REGISTRY ?= ghcr.io/mutablelogic

# Set docker tag
BUILD_TAG := ${DOCKER_REGISTRY}/go-whisper-${OS}-${ARCH}:${VERSION}
ROOT_PATH := $(CURDIR)

# Build docker container
docker: docker-dep submodule
	@echo build docker image: ${BUILD_TAG} for ${OS}/${ARCH}
	@${DOCKER} build \
		--tag ${BUILD_TAG} \
		--build-arg ARCH=${ARCH} \
		--build-arg OS=${OS} \
		--build-arg SOURCE=${BUILD_MODULE} \
		--build-arg VERSION=${VERSION} \
		-f etc/Dockerfile.${ARCH} .

# Test whisper bindings
test: libwhisper libggml
	@echo "Running tests"
	@CGO_CFLAGS="-I${ROOT_PATH}/third_party/whisper.cpp/include -I${ROOT_PATH}/third_party/whisper.cpp/ggml/include" \
	 CGO_LDFLAGS="-L${ROOT_PATH}/third_party/whisper.cpp" \
	 go test -v ./sys/whisper/...

# Build whisper-static-library
libwhisper: submodule
	@echo "Building libwhisper.a"
	@cd third_party/whisper.cpp && make -j$(nproc) libwhisper.a

# Build ggml-static-library
libggml: submodule
	@echo "Building libggml.a"
	@cd third_party/whisper.cpp && make -j$(nproc) libggml.a

# Build whisper-server
whisper-server: submodule
	@echo "Building whisper-server"
	@cd third_party/whisper.cpp && make -j$(nproc) server
	
# Push docker container
docker-push: docker-dep 
	@echo push docker image: ${BUILD_TAG}
	@${DOCKER} push ${BUILD_TAG}

# Update submodule to the latest version
submodule-update: git-dep
	@echo "Updating submodules"
	@${GIT} submodule foreach git pull origin master

# Submodule checkout
submodule: git-dep
	@echo "Checking out submodules"
	@${GIT} submodule update --init --recursive --remote

# Check for docker
docker-dep:
	@test -f "${DOCKER}" && test -x "${DOCKER}"  || (echo "Missing docker binary" && exit 1)

# Check for git
git-dep:
	@test -f "${GIT}" && test -x "${GIT}"  || (echo "Missing git binary" && exit 1)