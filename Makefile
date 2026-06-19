VERSION ?= dev
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"
DIST := dist

build: $(PLATFORMS)

$(PLATFORMS):
	@mkdir -p $(DIST)
	GOOS=$(word 1,$(subst /, ,$@)) GOARCH=$(word 2,$(subst /, ,$@)) \
		go build $(LDFLAGS) -o $(DIST)/cos-sync-$(subst /,-,$@)$(if $(filter windows/%,$@),.exe) .

clean:
	rm -rf $(DIST)

.PHONY: build clean $(PLATFORMS)
