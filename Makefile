BIN_NAME=ysmtp
BIN_DIR=bin

.PHONY: build build-docker run run-docker clean test

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(BIN_NAME) .

build-docker:
	docker build --network=host -t $(BIN_NAME) .

run-docker:
	docker run -p 25:25 $(BIN_NAME)

clean:
	rm -rf $(BIN_DIR)
