
build:
	docker run -d --name l0mywayDB -e POSTGRES_USER=dev -e POSTGRES_PASSWORD=pass -e PGUSER=dev -e POSTGRES_DB=l0myway -p 5433:5432 -d --rm postgres

run: build
	docker run -d --name natsjets -p 4222:4222 nats:alpine -js

migrate:
	migrate -path ./schema -database "postgres://dev:pass@localhost:5433/l0myway?sslmode=disable" up

migrate-down:
	migrate -path ./schema -database "postgres://dev:pass@localhost:5433/l0myway?sslmode=disable" down