.PHONY: test fuzz lint gen

lint:
	golangci-lint run -c .golangci.yml ./...
	@echo gopls check -severity hint ./...
	@RESULT=$$(grep -rL '^// Code generated .* DO NOT EDIT' --include="*.go" . | bash -c ' xargs gopls check -severity hint 2>&1'); \
	if [ -n "$$RESULT" ]; then \
		echo "$$RESULT"; \
		exit 1; \
	fi

bench:
	cd _benchmark/cmark && go run goldmark_benchmark.go 500 _data.md cpu.pprof && mv cpu.pprof ../../cpu.pprof

test:
	go test -coverprofile=profile.out -coverpkg=github.com/yuin/goldmark/v2,github.com/yuin/goldmark/v2/ast,github.com/yuin/goldmark/v2/extension,github.com/yuin/goldmark/v2/extension/ast,github.com/yuin/goldmark/v2/parser,github.com/yuin/goldmark/v2/renderer,github.com/yuin/goldmark/v2/renderer/html,github.com/yuin/goldmark/v2/text,github.com/yuin/goldmark/v2/util ./...

cov: test
	go tool cover -html=profile.out

fuzz:
	cd ./fuzz && go test -fuzz=FuzzDefault

gen:
	go generate ./...
