package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"testing"
)

var ctx context.Context = nil

var in = events.APIGatewayProxyRequest{
	Body: "{ " +
		"\"service\": \"https://graphical.weather.gov:443/xml/SOAP_server/ndfdXMLserver.php\", " +
		"\"requestBody\": \"https://graphical.weather.gov/xml/docs/SOAP_Requests/LatLonListZipCode.xml\", " +
		"\"responseMap\": { \"latLonList\": {} } }",
}

func TestHandler(t *testing.T) {
	out, err := Handler(ctx, in)
	if err != nil {
		t.Error("ERROR")
	}
	t.Log(out)
}
