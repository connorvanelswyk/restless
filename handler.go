package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
)

func Handler(ctx context.Context, in events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	sr, err := NewSoaplessRequest(in)
	if err != nil {
		return Error(err)
	}

	req, err := NewHttpRequest(*sr)
	if err != nil {
		return Error(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Error(err)
	}

	json, err := NewJsonResponseBody(*resp, *sr)
	if err != nil {
		return Error(err)
	}

	return Success(json)
}

func main() {
	lambda.Start(Handler)
}
