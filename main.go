package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"api-gateway.com/middleware"
	"api-gateway.com/proto/auth"
	"api-gateway.com/proto/encounters"
	follower "api-gateway.com/proto/followers"
	"api-gateway.com/proto/tours"
	"api-gateway.com/utils"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	gwmux := runtime.NewServeMux()

	authConnection, err := grpc.DialContext(
		context.Background(),
		"auth:90",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	defer authConnection.Close()

	followersConnection, err := grpc.DialContext(
		context.Background(),
		"followers:87",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer followersConnection.Close()

	tourConnection, err := grpc.DialContext(
		context.Background(),
		"tours:88",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer tourConnection.Close()

	encounterConnection, err := grpc.DialContext(
		context.Background(),
		"encounters:8082", //or localhost
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Println(err)
	}
	defer encounterConnection.Close()

	authClient := auth.NewAuthServiceClient(authConnection)
	err = auth.RegisterAuthServiceHandlerClient(context.Background(), gwmux, authClient)
	if err != nil {
		log.Fatalln(err)
	}

	tourClient := tours.NewToursServiceClient(tourConnection)
	err = tours.RegisterToursServiceHandlerClient(context.Background(), gwmux, tourClient)
	if err != nil {
		log.Fatalln(err)
	}

	followersClient := follower.NewFollowersServiceClient(followersConnection)
	err = follower.RegisterFollowersServiceHandlerClient(context.Background(), gwmux, followersClient)
	if err != nil {
		log.Fatalln(err)
	}

	encountersClient := encounters.NewEncountersServiceClient(encounterConnection)
	err = encounters.RegisterEncountersServiceHandlerClient(context.Background(), gwmux, encountersClient)
	if err != nil {
		fmt.Println(err)
	}

	handler := middleware.JwtMiddleware(gwmux, utils.GetProtectedPaths())
	gwServer := &http.Server{Addr: ":8083", Handler: handler}
	gwServer.Handler = addCorsMiddleware(gwmux)

	go func() {
		fmt.Println("Starting HTTP server on port 8083")
		if err := gwServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("ListenAndServe(): %v", err)
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGTERM, syscall.SIGINT)
	<-stopCh

	fmt.Println("Shutting down the server...")

	if err := gwServer.Shutdown(context.Background()); err != nil {
		log.Fatalln(err)
	}

	log.Println("Server gracefully stopped")
}

func addCorsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
