package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"main.go/handlers"
)

func main() {
	// Log string to standard out
	l := log.New(os.Stdout, "product-api", log.LstdFlags)

	// create the handlers from the seperate handler files and pass in the log
	hh := handlers.NewHello(l)
	hb := handlers.NewBye(l)

	// make a serve mux which takes the path and the handler func above
	sm := http.NewServeMux()
	sm.Handle("/", hh)
	sm.Handle("/bye", hb)

	// create a custom server - otherwise you could use: http.ListenAndServe(":9090", sm)
	s := &http.Server{
		Addr:         ":9090",
		Handler:      sm,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	// Start the server, this won't block so block with the channel below
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			l.Fatal(err)
		}
	}()

	//Make a channel that blocks (and keeps the server open)
	sigChan := make(chan os.Signal, 1)
	//Catch any kill calls like command c
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	//Handle Graceful shutdown, call shutdown with time out
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownRelease()

	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")

}
