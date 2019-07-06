default: build

build:
	go build -mod=readonly ./cmd/rcon-hub

test:
	go test -mod=readonly ./...

release:
	curl -sL https://git.io/goreleaser | bash

snapshot:
	goreleaser release --snapshot --rm-dist
