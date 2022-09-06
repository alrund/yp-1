
up:
	docker-compose up -d --build

down:
	docker-compose down --remove-orphans

lint:
	golangci-lint run

fix:
	golangci-lint run --fix

test:
	go test -v ./...

bench:
	go test -bench . ./...

benchmem:
	go test -bench . ./... -benchmem

race:
	go test -v -race ./...

cover:
	go test -short -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	rm coverage.out

cover-html:
	go test -short -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out