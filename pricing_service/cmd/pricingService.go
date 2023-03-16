package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	zero "github.com/rs/zerolog/log"

	"github.com/crypto-pricing-service/pricing_service/server"
)

const DefaultServerListenAddress = ":50443"

var apiKey, serverListenAddress string

func initFlags() {
	flag.StringVar(&apiKey, "api-key", "", "The api-key for coinmarketcap services ")
	flag.StringVar(&serverListenAddress, "server-address", DefaultServerListenAddress, "The the address the service is listening on")
	flag.Parse()
}

func main() {
	initFlags()

	if apiKey == "" {
		zero.Error().Msg("missing api key, please use --api-key <key> or -h for all options")
		return
	}

	s, err := server.NewGrpcServer(serverListenAddress, apiKey)
	if err != nil {
		zero.Fatal().Msgf("error:%v", err)
	}

	go s.StartGrpcServer()
	sCh := make(chan os.Signal)
	signal.Notify(sCh, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	select {
	case sig := <-sCh:
		zero.Error().Msgf("received OS signal %v, Shutting down gRPC server", sig)
		s.Stop()
	}
}
