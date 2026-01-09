SHELL := /bin/sh

.PHONY: dev-frontend dev-backend docker-up docker-down fmt

dev-frontend:
	cd frontend && npm install && npm run dev

dev-backend:
	cd backend && go run ./...

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down -v

fmt:
	cd backend && gofmt -w *.go
