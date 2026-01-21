# Paths to packages
DOCKER=$(shell which docker)
GIT=$(shell which git)
GO=$(shell which go)
CMAKE=$(shell which cmake)

# Set OS and Architecture
ARCH ?= $(shell arch | tr A-Z a-z | sed 's/x86_64/amd64/' | sed 's/i386/amd64/' | sed 's/armv7l/arm/' | sed 's/aarch64/arm64/')
OS ?= $(shell uname | tr A-Z a-z)
VERSION ?= $(shell git describe --tags --always | sed 's/^v//')

# Build parameters
ROOT_PATH := $(CURDIR)
BUILD_DIR ?= "build"
BUILD_JOBS ?= -j
PREFIX ?= ${BUILD_DIR}/install

# Build flags
BUILD_MODULE := $(shell cat go.mod | head -1 | cut -d ' ' -f 2)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitTag=$(shell git describe --tags --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/version.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 
CMAKE_FLAGS = -DBUILD_SHARED_LIBS=OFF

# Test flags
TEST_FLAGS ?=

# Default docker file is non-cuda
DOCKER_FILE := etc/Dockerfile.vulkan
DOCKER_SUFFIX := ""
DOCKER_REGISTRY ?= ghcr.io/mutablelogic

# If GGML_CUDA is set, then add a cuda tag for the go ${BUILD FLAGS}
# Target specific CUDA architectures
# https://developer.nvidia.com/cuda/gpus
ifeq ($(GGML_CUDA),1)
	TEST_FLAGS += -tags cuda
	BUILD_FLAGS += -tags cuda
	CMAKE_FLAGS += -DGGML_CUDA=ON
	BUILD_JOBS = -j2
	DOCKER_FILE = etc/Dockerfile.cuda
	DOCKER_SUFFIX = -cuda
	ifeq ($(ARCH),arm64)
		CMAKE_FLAGS += -DGGML_NATIVE=OFF '-DCMAKE_CUDA_ARCHITECTURES=87'
	endif
	ifeq ($(ARCH),amd64)
		CMAKE_FLAGS += -DGGML_NATIVE=OFF '-DCMAKE_CUDA_ARCHITECTURES=75;86;89'
	endif
endif

# If GGML_VULKAN is set, then add a vulkan tag for the go ${BUILD FLAGS}
ifeq ($(GGML_VULKAN),1)
	TEST_FLAGS += -tags vulkan 
	BUILD_FLAGS += -tags vulkan
	CMAKE_FLAGS += -DGGML_VULKAN=ON -DGGML_NATIVE=OFF
	DOCKER_FILE = etc/Dockerfile.vulkan
endif

# Docker
DOCKER_TAG := ${DOCKER_REGISTRY}/go-whisper${DOCKER_SUFFIX}-${OS}-${ARCH}:${VERSION}

# Targets
all: gowhisper

#####################################################################
# BUILD

# Make gowhisper (includes server run command)
gowhisper: generate libwhisper libffmpeg
	@echo "Building gowhisper"
	@PKG_CONFIG_PATH=$(shell realpath ${PREFIX})/lib/pkgconfig CGO_LDFLAGS_ALLOW="-(W|D).*" ${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/gowhisper ./cmd/gowhisper

# Make gowhisper-client (no server run command)
gowhisper-client: 
	@echo "Building gowhisper-client"
	@${GO} build ${BUILD_FLAGS} -tags client -o ${BUILD_DIR}/gowhisper ./cmd/gowhisper

# Generate the pkg-config files
generate: mkdir go-tidy libwhisper
	@echo "Generating pkg-config"
	@mkdir -p ${BUILD_DIR}/lib/pkgconfig
	@PKG_CONFIG_PATH=$(shell realpath ${PREFIX})/lib/pkgconfig PREFIX="$(shell realpath ${PREFIX})" go generate ./sys/whisper

# make libwhisper and install at ${PREFIX}
libwhisper: mkdir submodule cmake-dep 
	@echo "Making libwhisper with ${CMAKE_FLAGS}"
	@${CMAKE} -S third_party/whisper.cpp -B ${BUILD_DIR} -DCMAKE_BUILD_TYPE=Release ${CMAKE_FLAGS}
	@${CMAKE} --build ${BUILD_DIR} ${BUILD_JOBS} --config Release
	@${CMAKE} --install ${BUILD_DIR} --prefix $(shell realpath ${PREFIX})

# make ffmpeg libraries and install at ${PREFIX}
libffmpeg: mkdir submodule
	@echo "Making ffmpeg libraries => ${PREFIX}"
	@mkdir -p ${BUILD_DIR}
	@mkdir -p ${PREFIX}
	@BUILD_DIR=$(shell realpath ${BUILD_DIR}) PREFIX=$(shell realpath ${PREFIX}) make -C third_party/go-media ffmpeg

#####################################################################
# TEST

# Test whisper
test: test-sys test-pkg

# Test whisper pkg bindings
test-pkg: generate libwhisper libffmpeg
	@echo "Running tests (pkg)"
	@PKG_CONFIG_PATH=$(shell realpath ${PREFIX})/lib/pkgconfig ${GO} test ${TEST_FLAGS} ./pkg/...

# Test whisper bindings
test-sys: generate libwhisper
	@echo "Running tests (sys) with ${PREFIX}/lib"
	@PKG_CONFIG_PATH=$(shell realpath ${PREFIX})/lib/pkgconfig ${GO} test ${TEST_FLAGS} ./sys/whisper/...

#####################################################################
# DOCKER

# Build docker container
docker: docker-dep submodule
	@echo build docker image: ${DOCKER_TAG} for ${OS}/${ARCH}
	@${DOCKER} build \
		--tag ${DOCKER_TAG} \
		--build-arg ARCH=${ARCH} \
		--build-arg OS=${OS} \
		--build-arg SOURCE=${BUILD_MODULE} \
		--build-arg VERSION=${VERSION} \
		--build-arg GGML_CUDA=${GGML_CUDA} \
		--build-arg GGML_VULKAN=${GGML_VULKAN} \
		-f ${DOCKER_FILE} .

# Push docker container
docker-push: docker-dep 
	@echo push docker image: ${DOCKER_TAG}
	@${DOCKER} push ${DOCKER_TAG}

#####################################################################
# THIRD PARTY DEPENDENCIES

# Update submodule to the latest version
submodule-update: git-dep
	@echo "Updating submodules"
	@${GIT} submodule foreach git pull origin master

# Submodule checkout
submodule: git-dep
	@echo "Checking out submodules"
	@${GIT} submodule update --init --recursive --remote

# Submodule clean (ONLY cleans submodules, not main repo)
submodule-clean: git-dep
	@echo "Cleaning submodules only"
	@${GIT} submodule sync --recursive
	@${GIT} submodule update --init --force --recursive
	@${GIT} submodule foreach --recursive git clean -ffdx	

# Check for docker
docker-dep:
	@test -f "${DOCKER}" && test -x "${DOCKER}"  || (echo "Missing docker binary" && exit 1)

# Check for docker
cmake-dep:
	@test -f "${CMAKE}" && test -x "${CMAKE}"  || (echo "Missing cmake binary" && exit 1)

# Check for git
git-dep:
	@test -f "${GIT}" && test -x "${GIT}"  || (echo "Missing git binary" && exit 1)

# Check for go
go-dep:
	@test -f "${GO}" && test -x "${GO}"  || (echo "Missing go binary" && exit 1)

#####################################################################
# CLEAN

# Make build directory
mkdir:
	@echo Mkdir ${BUILD_DIR}
	@install -d ${BUILD_DIR}
	@echo Mkdir ${PREFIX}
	@install -d ${PREFIX}

# go mod tidy
go-tidy: go-dep
	@echo Tidy
	@${GO} mod tidy
	@${GO} clean -cache

# Clean - only removes build artifacts, does NOT reset git or clean submodules
clean:
	@echo "Cleaning build artifacts"
	@rm -rf ${BUILD_DIR}
	@${GO} clean -cache
	@test -f third_party/go-media/Makefile && make -C third_party/go-media clean || true
