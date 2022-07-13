
up:
	docker-compose up -d --build

down:
	docker-compose down --remove-orphans

lint:
	golangci-lint run

fix:
	golangci-lint run --fix

