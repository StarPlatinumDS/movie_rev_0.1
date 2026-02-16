run:
	docker-compose up -d
	@go run ./cmd/reviews
	
down:
	docker-compose down
