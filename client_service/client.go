package main

import (
	"context"
	"flag"
	"runtime"
	"sync"
	"time"

	zero "github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	pb "github.com/crypto-pricing-service/pricing_service/grpc"
)

const DefaultServerListenAddress = ":50443"

var (
	coins      sync.Map
	coinStatus status
)

type grpcClient struct {
	conn pb.PricingServiceClient
}

type status struct {
	last    float64
	current float64
	change  float64
}

func (c *grpcClient) Close() error {
	return c.Close()
}

// newGrpcClient create new gRPC client that connects to pricing server
func newGrpcClient(addr string) (*grpcClient, error) {
	cert := grpc.WithTransportCredentials(insecure.NewCredentials())

	keepaliveParams := grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	})
	conn, err := grpc.Dial(addr, cert, keepaliveParams, grpc.WithBlock())
	return &grpcClient{pb.NewPricingServiceClient(conn)}, err
}

// getCoinValue make a gRPC call to the pricing server, sends a symbol and receives current price
func (c *grpcClient) getCoinValue(coin string) (float64, error) {
	response, err := c.conn.GetPrices(context.Background(), &pb.Request{TokenName: coin})
	if err != nil {
		zero.Error().Msgf("getCoinValue failed:%v", err)
		return 0, err
	}
	return response.Price, err
}

// checkCoinsValues get the coins status and prints to console
func (c *grpcClient) checkCoinsValues() {
	for {
		coins.Range(func(k interface{}, v interface{}) bool {
			go func(coin string, value float64) {
				newPrice, err := c.getCoinValue(coin)
				if err != nil {
					zero.Error().Msgf("checkCoinsValues failed:%v for coin:&s", err, coin)
				}
				coins.Store(coin, newPrice)
				coinStatus = status{
					last:    value,
					current: newPrice,
					change:  (1 - (value / newPrice)) * 100,
				}
				zero.Info().Msgf("%s: %f, %f, %.2f%%", coin, coinStatus.last, coinStatus.current, coinStatus.change)
			}(k.(string), v.(float64))
			return true
		})
		time.Sleep(time.Minute)
	}
}

// initCoinsSyncMap init coins sync map with the desired symbols
func initCoinsSyncMap() {
	coins.Store("BTC", 0.0)
	coins.Store("ETH", 0.0)
	coins.Store("USDT", 0.0)
}

func main() {
	clientListenAddress := flag.String("server-address", DefaultServerListenAddress, "The the address the service is listening on")
	flag.Parse()

	zero.Info().Msgf("gRPC client starting listening on:%s", *clientListenAddress)
	initCoinsSyncMap()
	c, err := newGrpcClient(*clientListenAddress)
	if err != nil {
		zero.Fatal().Msgf("failed starting gRPC client:%v", err)
	}
	c.checkCoinsValues()
	runtime.Goexit()
}
