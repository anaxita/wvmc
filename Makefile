.PHONY b:
b:
	# go build -v -o ./build/wvmc.exe ./cmd/wvmc
	go build -v ./cmd/wvmc

.PHONY bc:
bc:
	go build -v -ldflags="-s -w -H windowsgui" -o ./wvmc.exe ./cmd/wvmc

.PHONY run:
run:
	go run ./cmd/wvmc/main.go

.PHONY test:
test:
	go test ./...

.DEFAULT_GOAL := run