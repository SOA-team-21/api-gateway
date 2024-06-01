package main

import (
	"context"
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
		"localhost:90",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	defer authConnection.Close()

	followersConnection, err := grpc.DialContext(
		context.Background(),
		"localhost:8080",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer followersConnection.Close()

	tourConnection, err := grpc.DialContext(
		context.Background(),
		"localhost:88",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer tourConnection.Close()

	encounterConnection, err := grpc.DialContext(
		context.Background(),
		"localhost:8082",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln(err)
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
		log.Fatalln(err)
	}

	handler := middleware.JwtMiddleware(gwmux, utils.GetProtectedPaths())
	gwServer := &http.Server{Addr: ":8083", Handler: handler}
	gwServer.Handler = addCorsMiddleware(gwmux)
	go func() {
		log.Println("Starting HTTP server on port 8083")
		if err := gwServer.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, syscall.SIGTERM)
	<-stopCh

	if err = gwServer.Close(); err != nil {
		log.Fatalln(err)
	}

}

func addCorsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		//w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
