# Helper makefile to build and run the docker container. It assumes that you are
# a member of the docker group.

TAG := tsatusbii-auth:0.1
UID := `id -u`
GID := `id -g`
BUILD_ARGS = --build-arg UID=$(UID) --build-arg GID=$(GID)
SRC = $(wildcard *.go)
TPL = $(wildcard tpl/*.gohtml)
CONFIG = config.toml

all: docker-build 

.PHONY: docker-build
docker-build: $(SRC) $(TPL) $(CONFIG)
	docker build $(BUILD_ARGS) -t $(TAG) .

.PHONY: run
run: docker-build
	docker run -it --publish 8000:8000 $(TAG)

.PHONY: clean
clean:
	docker rm --force $(TAG)
