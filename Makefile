.PHONY: all

SRCS = configuration.go httpd.go server.go status_response_writer.go config_fetcher.go
VERSION = $(shell git rev-parse HEAD)
CODENAME = "Soyouz"
DATE = $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
REPO_OWNER = "wingz-inc"
REPO_NAME = "httpd"

dockerimage: httpd
	docker build -t httpd:latest .

darwin:
	GOARCH=amd64 GOOS=darwin CGO_ENABLED=0 go build -ldflags "-s -w -X github.com/$(REPO_OWNER)/$(REPO_NAME)/version.Version=$(VERSION) -X github.com/$(REPO_OWNER)/$(REPO_NAME)/version.Codename=$(CODENAME) -X github.com/$(REPO_OWNER)/$(REPO_NAME)/version.BuildDate=$(DATE)" -o "httpd" .

all: dockerimage

httpd: $(SRCS)
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -ldflags "-s -w -X github.com/$(REPO_OWNER)/$(REPO_NAME)/version.version=$(VERSION) -X github.com/$(REPO_OWNER)/$(REPO_NAME)/main.version.Codename=$(CODENAME) -X github.com/$(REPO_OWNER)/$(REPO_NAME)/version.BuildDate=$(DATE)" -o "httpd" .

clean:
	rm httpd
