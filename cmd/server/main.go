package main

import (
	"canvas-report/api"
	"canvas-report/canvas"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	canvasBaseUrl := os.Getenv("CANVAS_BASE_URL")
	if canvasBaseUrl == "" {
		panic("missing env: CANVAS_BASE_URL")
	}

	canvasAccessToken := os.Getenv("CANVAS_ACCESS_TOKEN")
	if canvasAccessToken == "" {
		panic("missing env: CANVAS_ACCESS_TOKEN")
	}

	// Number of items to fetch per request when paginating with the Canvas API.
	pageSize := 100

	pageSizeEnv := os.Getenv("CANVAS_PAGE_SIZE")
	if pageSizeEnv != "" {
		value, err := strconv.Atoi(pageSizeEnv)
		if err != nil {
			panic("invalid env: CANVAS_PAGE_SIZE")
		}

		pageSize = value
	}

	canvasClient, err := canvas.NewCanvasClient(canvasBaseUrl, canvasAccessToken, pageSize)
	if err != nil {
		panic(fmt.Errorf("error creating canvas client: %w", err))
	}

	apiController, err := api.NewAPIController(canvasClient, nil)
	if err != nil {
		panic(fmt.Errorf("error creating api controller: %w", err))
	}

	router := api.NewRouter(apiController, nil)

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: router,
	}

	go func() {
		log.Println("staring server...")

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("error listen and serve: %s\n", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	signal := <-signalChan

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	log.Printf("received signal %s, shutting down sever...\n", signal)

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("error shutting down sever: %s\n", err)
	}
}
