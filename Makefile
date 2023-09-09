-include .env

lint:
	golangci-lint run

# install a migrate tool
migrate-install:
	go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.2

# create new migrations
migrate-new:
	migrate create -ext sql -dir ./migrations "$(name)"

LOCAL_DB_DSN = "sqlite3://${SQLITE_NAME}?query&_auth&_auth_user=${SQLITE_USER}&_auth_pass=${SQLITE_PASSWORD}&_auth_crypt=sha512"

# up migrations
migrate-up:
	migrate -path ./migrations -database ${LOCAL_DB_DSN} up

# rollback the last migration
migrate-down:
	migrate -path ./migrations -database ${LOCAL_DB_DSN} down

# generate
deps:
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.12.4 # OpenAPI
	go install github.com/golang/mock/mockgen@v1.6.0 # Mock

gen:
	go generate ./...
#	docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate -i /local/OpenApi.yml -g go-server -o /local/internal/api/gen
#	oapi-codegen -generate gorilla -package gen OpenApi.yml > internal/api/gen/server.go
#	oapi-codegen -generate types -package gen OpenApi.yml > internal/api/gen/types.go
#	mockgen -source=internal/service/bonuses.go -destination=../tests/mocks.go -package=tests

# tests
tests:
	go test -v -race -count 1 ./...