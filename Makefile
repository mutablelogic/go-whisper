# Paths to tools needed in dependencies
GO := $(shell which go)
GIT := $(shell which git)
CMAKE := $(shell which cmake)

# Build flags
BUILD_MODULE := $(shell go list -m)
BUILD_FLAGS = -ldflags "-s -w" 

# Paths to locations, etc
BUILD_DIR := "build"
MODEL_DIR := "models"
CMD_DIR := $(wildcard cmd/*)

# Targets
all: clean whisper cmd

submodule:
	@echo Update submodules
	@${GIT} submodule update --init --recursive

whisper: submodule
	@echo Build whisper
	@${CMAKE} -S third_party/whisper.cpp -B ${BUILD_DIR} -D BUILD_SHARED_LIBS=off
	@${CMAKE} --build ${BUILD_DIR} --target whisper

models: model-tiny model-small model-medium model-large

model-tiny: mkdir whisper
	@echo Download model-tiny
	@third_party/whisper.cpp/models/download-ggml-model.sh tiny
	@install third_party/whisper.cpp/models/ggml-tiny.bin ${MODEL_DIR}/ggml-tiny.bin

model-small: mkdir whisper
	@echo Download model-small
	@third_party/whisper.cpp/models/download-ggml-model.sh small
	@install third_party/whisper.cpp/models/ggml-small.bin ${MODEL_DIR}/ggml-small.bin

model-medium: mkdir whisper
	@echo Download model-medium
	@third_party/whisper.cpp/models/download-ggml-model.sh medium
	@install third_party/whisper.cpp/models/ggml-medium.bin ${MODEL_DIR}/ggml-medium.bin

model-large: mkdir whisper
	@echo Download model-large
	@third_party/whisper.cpp/models/download-ggml-model.sh large
	@install third_party/whisper.cpp/models/ggml-large.bin ${MODEL_DIR}/ggml-large.bin

cmd: $(filter-out cmd/README.md, $(wildcard cmd/*))

test: whisper
	@${GO} mod tidy
	@${GO} test -v ./sys/...

$(CMD_DIR): dependencies mkdir
	@echo Build cmd $(notdir $@)
	@${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@) ./$@

FORCE:

dependencies:
	@test -x ${GO} || (echo "Missing go binary" && exit 1)

mkdir:
	@echo Mkdir ${BUILD_DIR} ${MODEL_DIR}
	@install -d ${BUILD_DIR}
	@install -d ${MODEL_DIR}

clean:
	@echo Clean
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean

