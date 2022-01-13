.PHONY build:
build:
	GOOS=windows GOARCH=arm64 go build -v -ldflags="-s -w -H windowsgui" -o ./wvmc.exe ./cmd/wvmc

.PHONY build-dev:
build-dev:
	GOOS=windows GOARCH=arm64 go build -v -ldflags="-s -w -H windowsgui" -o ./wvmc_dev.exe ./cmd/wvmc

.PHONY run:
run:
	go run ./cmd/wvmc/main.go

.PHONY test:
test:
	go test ./...

.DEFAULT_GOAL := run