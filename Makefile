
migrate-up:
	migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/cozybox?sslmode=disable" up

migrate-down:
	migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/cozybox?sslmode=disable" down

migrate-down-force:
	migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/cozybox?sslmode=disable" down

.PHONY: migrate-up migrate-down migrate-down-force
