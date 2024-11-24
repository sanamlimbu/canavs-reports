package main

import (
	"canvas-report/api"
	"canvas-report/canvas"
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
)

var chiLambda *chiadapter.ChiLambda

func init() {
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

	chiLambda = chiadapter.New(router)
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return chiLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(handler)
}
