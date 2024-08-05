queries:
	-rm db/*.go
	sqlc generate
	

reset-postgres:
	docker stop postgres
	docker rm postgres
	docker run --name postgres -e POSTGRES_PASSWORD=postgres -d -p 5432:5432 postgres

reset-redpanda:
	docker compose down -v
	docker compose up -d

pause:
	sleep 5

schema:
	migrate -database "postgresql://postgres:postgres@localhost:5432?sslmode=disable" -path=db/migrations up

db-all: reset-postgres pause schema queries reset-redpanda
	@echo done.

cli:
	PGPASSWORD=postgres psql -h localhost -p 5432 -U postgres

