package main

import (
	"./usecase"
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(ctx context.Context, in events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return usecase.Handle(in)
}

func main() {
	lambda.Start(Handler)
}
