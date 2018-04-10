PROJECT=vips-go

.PHONY: format
format:
	@gofmt -w -s .

.PHONY: lint
lint:
	@go get golang.org/x/lint/golint
	@go get honnef.co/go/tools/cmd/megacheck
	@golint vips/...
	@megacheck ./...

.PHONY: test
test:
	@go test ./...

.PHONY: name
name:
	@echo $(PROJECT)
