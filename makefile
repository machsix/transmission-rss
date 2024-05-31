MODULE := github.com/Asutorufa/transmission-rss

CGO_ENABLED := 0

GO=$(shell command -v go | head -n1)


GO_BUILD_CMD=CGO_ENABLED=$(CGO_ENABLED) $(GO) build -trimpath

.PHONY: transmission-rss
transmission-rss:
	$(GO_BUILD_CMD) -v .