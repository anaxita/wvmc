.PHONY b:
b:
	go build -v -o ./build/wvmc.exe ./cmd/wvmc

.PHONY bc:
bc:
	go build -v -ldflags="-s -w" -o ./build/wvmc.exe ./cmd/wvmc

.PHONY run:
run:
	go run ./cmd/wvmc

.PHONY test:
test:
	go test ./...

.DEFAULT_GOAL := run