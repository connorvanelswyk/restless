package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/beevik/etree"
	"html"
	"io"
	"log"
	"net/http"
	"strings"
)

type requester struct {
	ServiceUrl        string            `json:"serviceUrl"` // The endpoint for the desired SOAP service
	RequestUrl        string            `json:"requestUrl"` // The location of an empty or sample SOAP service request
	RequestMethod     string            `json:"requestMethod"`
	RequestProperties map[string]string `json:"requestProperties"`
	RequestMap        map[string]string `json:"requestMap"`
	ResponseMap       map[string]string `json:"responseMap"`
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// unmarshal the proxy request body
	r := requester{}
	err := json.Unmarshal([]byte(req.Body), &r)
	if err != nil {
		panic(err.Error())
	}

	// get the example service request url
	resp, err := http.Get(r.RequestUrl)
	if err != nil {
		panic(err.Error())
	}

	// create xml from example service request body
	reqXml := etree.NewDocument()
	_, err = reqXml.ReadFrom(resp.Body)

	updateRequestXml(r.RequestMap, *reqXml)

	// write it to a string for request body
	s, err := reqXml.WriteToString()
	if err != nil {
		panic(err.Error())
	}

	// create the service request and set the request headers
	request, err := http.NewRequest(r.RequestMethod, r.ServiceUrl, strings.NewReader(s))
	if err != nil {
		panic(err.Error())
	}
	for k, v := range r.RequestProperties {
		request.Header.Add(k, v)
	}

	// create client and get response from service
	client := &http.Client{}
	resp, err = client.Do(request)
	if err != nil {
		panic(err.Error())
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			panic(err.Error())
		}
	default:
		reader = resp.Body
	}

	// marshal response body bytes into xml string where html is unescaped
	respXml := etree.NewDocument()
	_, err = respXml.ReadFrom(reader)
	if err != nil {
		panic(err.Error())
	}
	s, err = respXml.WriteToString()
	if err != nil {
		panic(err.Error())
	}

	body := createResponseBodyJson(html.UnescapeString(s), r.ResponseMap)

	returnVal := events.APIGatewayProxyResponse{
		Body:            body,
		StatusCode:      200,
		IsBase64Encoded: false,
	}

	return returnVal, nil
}

//  update request xml (d) with properties
//    from request map (m)
//   where request map keys are element tag names
//     and request map values are element text values
func updateRequestXml(m map[string]string, d etree.Document) {
	for k, v := range m {
		e := d.FindElement("//" + k)
		e.SetText(v)
	}
}

func createResponseBodyJson(html string, responseMap map[string]string) (jsonString string) {
	doc := etree.NewDocument()
	_, _ = doc.ReadFrom(strings.NewReader(html))
	log.Println(html)
	// create json map from parsing doc element text
	m := make(map[string]string)
	for k := range responseMap {
		e := doc.FindElement("//" + k)
		m[k] = e.Text()
	}

	// marshal map into json string for response body
	j, _ := json.Marshal(m)
	return string(j)
}

func main() {
	lambda.Start(Handler)
}
