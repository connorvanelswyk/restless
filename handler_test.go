package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"testing"
)

var ctx context.Context = nil

var in = events.APIGatewayProxyRequest{
	Body: "{\"service\": \"https://graphical.weather.gov/xml/SOAP_server/ndfdXMLserver.php\", \"requestBody\": \"https://graphical.weather.gov/xml/docs/SOAP_Requests/GmlLatLonList.xml\", \"requestMap\": {  \"requestedTime\": \"2019-06-22T23:59:59\"  },  \"responseMap\": {    \"gml:boundedBy\": {      \"gml:coordinates\": \"\"    },    \"gml:featureMember\": {      \"gml:coordinates\": \"\",      \"app:validTime\": \"\",      \"app:maximumTemperature\": \"\"    }  }}",
}

func TestHandler(t *testing.T) {
	_, err := Handler(ctx, in)
	if err != nil {
		t.Error("ERROR")
	}
}
