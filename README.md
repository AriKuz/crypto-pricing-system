# crypto-pricing-system

this system is designed to fetch USD prices for each of the following tokens - (BTC, ETH, USDT), every minute from
Coinmarketcap API.

### pricing_service
responsible for fetching USD prices for 3 tokens (BTC, ETH, USDT) every minute from Coinmarketcap API and storing them in local in memory cache.
it includes the following:
- fetcher: simple rest client to fetch data from Coinmarketcap API
- grpc: includes the protobuf data needed for gRPC
- server: runs a gRPC server on a chosen port (default is 50443)

### client_service
includes a gRPC client that communicates with pricing_service and gets the USD prices for the desired tokens.

### setup
clone the repository and run "make build", it will sync dependencies and compile the binaries for the 2 services.

### notes
when running server binary, you must pass the api-key as an argument:
./server --api-key "api-key"

both server and client accept a port to listen to, but if nothing is passed, then default is ":50443"

If for any reason you would like to re-compile the grpc files, run "make build" and then:

protoc --go_out=pricing_service/grpc --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=pricing_service/grpc pricing_service/grpc/data.proto
