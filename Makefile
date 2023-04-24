-include .env

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

# generate mock
gen:
	go generate ./...

