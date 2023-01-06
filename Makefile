# Paths to tools needed in dependencies
GO := $(shell which go)
GIT := $(shell which git)

# Build flags
BUILD_MODULE := $(shell go list -m)
BUILD_FLAGS = -ldflags "-s -w" 

# Paths to locations, etc
BUILD_DIR := build
MODEL_DIR := models
CMD_DIR := $(wildcard cmd/*)
INCLUDE_PATH := $(abspath third_party/whisper.cpp)
LIBRARY_PATH := $(abspath third_party/whisper.cpp)

# Targets
all: clean whisper cmd

submodule:
	@echo Update submodules
	@${GIT} submodule update --init --recursive --remote --force

whisper: submodule
	@echo Build whisper
	@make -C third_party/whisper.cpp libwhisper.a

model-downloader: submodule mkdir
	@echo Build model-downloader
	@make -C third_party/whisper.cpp/bindings/go examples/go-model-download
	@install third_party/whisper.cpp/bindings/go/build/go-model-download ${BUILD_DIR}

models: model-downloader
	@echo Downloading models
	@${BUILD_DIR}/go-model-download -out ${MODEL_DIR}

cmd: whisper $(wildcard cmd/*)

$(CMD_DIR): dependencies mkdir
	@echo Build cmd $(notdir $@)
	@C_INCLUDE_PATH=${INCLUDE_PATH} LIBRARY_PATH=${LIBRARY_PATH} ${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@) ./$@

FORCE:

dependencies:
	@test -f "${GO}" && test -x "${GO}"  || (echo "Missing go binary" && exit 1)
	@test -f "${GIT}" && test -x "${GIT}"  || (echo "Missing git binary" && exit 1)

mkdir:
	@echo Mkdir ${BUILD_DIR} ${MODEL_DIR}
	@install -d ${BUILD_DIR}
	@install -d ${MODEL_DIR}

clean:
	@echo Clean
	@rm -fr $(BUILD_DIR)
	@${GIT} submodule deinit --all -f
	@${GO} mod tidy
	@${GO} clean

