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

	// read the example service response body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	// create xml from example service request body
	reqXml := etree.NewDocument()
	_, err = reqXml.ReadFrom(getIsoDecodedReader(data))

	updateRequestXml(r.ResponseMap, *reqXml)

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

	// read response body into bytes
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	// marshal response body bytes into xml string
	respXml := etree.NewDocument()
	_, err = respXml.ReadFrom(getIsoDecodedReader(bodyBytes))
	s, err = respXml.WriteToString()

	// create doc from response xml where html is unescaped
	doc := etree.NewDocument()
	_, err = doc.ReadFrom(strings.NewReader(html.UnescapeString(s)))

	// create json map from parsing doc element text
	m := make(map[string]string)
	for k := range r.ResponseMap {
		e := doc.FindElement("//" + k)
		m[k] = e.Text()
		// todo - support lists and maps
	}

	// marshal map into json string for response body
	j, err := json.Marshal(m)
	returnVal := events.APIGatewayProxyResponse{
		Body:            string(j),
		StatusCode:      200,
		IsBase64Encoded: false,
	}

	return returnVal, nil
}

// where b is ISO-8859-1
//   and r is UTF-8
func getIsoDecodedReader(b []byte) (r io.Reader) {
	return charmap.ISO8859_1.NewDecoder().Reader(bytes.NewReader(b))
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

func main() {
	lambda.Start(Handler)
}
