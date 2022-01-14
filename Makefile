.PHONY build-prod:
build-prod:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H windowsgui" -o ./build/prod/wvmc.exe ./cmd/wvmc

.PHONY build-dev:
build-dev:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H windowsgui" -o ./build/dev/wvmc_dev.exe ./cmd/wvmc

.PHONY run:
run:
	go run ./cmd/wvmc/main.go

.PHONY test:
test:
	go test ./...

.DEFAULT_GOAL := run