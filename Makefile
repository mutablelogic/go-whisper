# Paths to tools needed in dependencies
GO := $(shell which go)
GIT := $(shell which git)
CMAKE := $(shell which cmake)

# Build flags
BUILD_MODULE := $(shell go list -m)
BUILD_FLAGS = -ldflags "-s -w" 

# Paths to locations, etc
BUILD_DIR := "build"
CMD_DIR := $(wildcard cmd/*)

# Targets
all: clean whisper cmd

submodule:
	@echo Update submodules
	@${GIT} submodule update --init --recursive

whisper: submodule
	@echo Build whisper
	${CMAKE} -S third_party/whisper.cpp -B ${BUILD_DIR}
	${CMAKE} --build ${BUILD_DIR} --target whisper

cmd: $(filter-out cmd/README.md, $(wildcard cmd/*))

test:
	@${GO} mod tidy
	@${GO} test -v ./pkg/...

$(CMD_DIR): dependencies mkdir
	@echo Build cmd $(notdir $@)
	@${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@) ./$@

FORCE:

dependencies:
	@test -x ${GO} || (echo "Missing go binary" && exit 1)

mkdir:
	@echo Mkdir ${BUILD_DIR}
	@install -d ${BUILD_DIR}

clean:
	@echo Clean
	@rm -fr $(BUILD_DIR)
	@${GIT} submodule deinit -f --all
	@${GO} mod tidy
	@${GO} clean

