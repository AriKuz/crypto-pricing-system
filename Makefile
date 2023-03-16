all: build

build: deps
	go build -o server pricing_service/cmd/pricingService.go
	go build -o client client_service/client.go

deps:
	go mod tidy

clean:
	rm -f server client
