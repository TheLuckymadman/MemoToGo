migrate create -ext sql -dir . memotogo_tables

migrate -path . -database "postgres://admin:admin@localhost:5432/admin?sslmode=disable&search_path=memotogo" force 20260313130202

migrate -path . -database "postgres://admin:admin@localhost:5432/admin?sslmode=disable&search_path=memotogo" up