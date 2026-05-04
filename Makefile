.PHONY: run doctor test build install clean paths release

APP := aether
VERSION := 1.0.0
CMD := ./cmd/aether
BIN_DIR := bin
DIST_DIR := dist
BIN := $(BIN_DIR)/$(APP)

run:
	go run $(CMD)

doctor:
	go run $(CMD) doctor

test:
	go test ./...

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) $(CMD)

install:
	go install $(CMD)

release: test
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)/$(APP)_$(VERSION)_linux_amd64
	GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(APP)_$(VERSION)_linux_amd64/$(APP) $(CMD)
	cp README.md README.ru.md LICENSE CHANGELOG.md $(DIST_DIR)/$(APP)_$(VERSION)_linux_amd64/
	cp -R assets $(DIST_DIR)/$(APP)_$(VERSION)_linux_amd64/
	tar -C $(DIST_DIR) -czf $(DIST_DIR)/$(APP)_$(VERSION)_linux_amd64.tar.gz $(APP)_$(VERSION)_linux_amd64

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR)

paths:
	go run $(CMD) config path
	go run $(CMD) history path
