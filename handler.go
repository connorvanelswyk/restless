package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/beevik/etree"
	"golang.org/x/text/encoding/charmap"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const willkommen = "Event handling commenced."
const tschuss = "Event handling concluded."
const unmarshalE = "Error unmarshalling body."
const loadingE = "Error loading request."
const bodyErr = "Error reading response body."
const reqUrlE = "Error reading request Url."

type requester struct {
	ServiceUrl        string            `json:"serviceUrl"` // The endpoint for the desired SOAP service
	RequestUrl        string            `json:"requestUrl"` // The location of an empty or sample SOAP service request
	RequestMethod     string            `json:"requestMethod"`
	RequestProperties map[string]string `json:"requestProperties"`
	RequestMap        map[string]string `json:"requestMap"`
	ResponseMap       map[string]string `json:"responseMap"`
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	log.Print(willkommen)

	returnVal := events.APIGatewayProxyResponse{
		Body:       "",
		StatusCode: 200,
	}

	r := requester{}
	err := json.Unmarshal([]byte(req.Body), &r)
	if err != nil {
		log.Print(unmarshalE)
	}

	// get the example request
	resp, err := http.Get(r.RequestUrl)
	if err != nil {
		log.Print(reqUrlE)
		panic(err.Error())
	}

	// read response body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(loadingE)
	}

	// create doc from reader
	doc := etree.NewDocument()
	_, err = doc.ReadFrom(getIsoDecodedReader(data))

	// set a different value
	// todo - read map from requester
	e := doc.FindElement("//listZipCodeList")
	e.SetText("33401")

	// write it to a string for request body
	s, err := doc.WriteToString()

	// create the request
	request, err := http.NewRequest(r.RequestMethod, r.ServiceUrl, strings.NewReader(s))
	if err != nil {
		panic(err.Error())
	}

	// set request headers
	client := &http.Client{}
	for k, v := range r.RequestProperties {
		request.Header.Add(k, v)
	}

	// get the response
	resp, err = client.Do(request)
	if err != nil {
		panic(err.Error())
	}

	// read response body
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(bodyErr)
	}

	responseDoc := etree.NewDocument()
	_, err = responseDoc.ReadFrom(getIsoDecodedReader(bodyBytes))

	// write it to a string for request body
	s, err = responseDoc.WriteToString()
	s = html.UnescapeString(s)

	// todo - read map from requester

	returnVal.Body = s

	log.Print(tschuss)

	return returnVal, nil
}

// convert from ISO-8859-1 to native UTF-8
func getIsoDecodedReader(b []byte) (reader io.Reader) {
	return charmap.ISO8859_1.NewDecoder().Reader(bytes.NewReader(b))
}

func main() {
	lambda.Start(Handler)
}
