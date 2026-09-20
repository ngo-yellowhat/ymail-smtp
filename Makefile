BIN_NAME=ysmtp
BIN_DIR=bin

.PHONY: build build-docker run run-docker clean test

build:
	go build -o $(BIN_DIR)/$(BIN_NAME) .

build-docker:
	docker build --network=host -t $(BIN_NAME) .

run: build
	./$(BIN_DIR)/$(BIN_NAME) start -p 2525

run-docker:
	docker run -p 2525:2525 $(BIN_NAME)

clean:
	rm -rf $(BIN_DIR)
