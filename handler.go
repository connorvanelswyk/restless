package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/antchfx/xmlquery"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/beevik/etree"
	"io"
	"net/http"
	"strings"
)

type requester struct {
	ServiceUrl        string                       `json:"serviceUrl"` // The endpoint for the desired SOAP service
	RequestUrl        string                       `json:"requestUrl"` // The location of an empty or sample SOAP service request
	RequestMethod     string                       `json:"requestMethod"`
	RequestProperties map[string]string            `json:"requestProperties"`
	RequestMap        map[string]string            `json:"requestMap"`
	ResponseMap       map[string]map[string]string `json:"responseMap"`
}

// M is an alias for map[string]interface{}
type M map[string]interface{}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// unmarshal the proxy request body
	r := requester{}
	err := json.Unmarshal([]byte(req.Body), &r)
	if err != nil {
		return newBlankAPIGatewayProxyResponse(), err
	}

	// get the example service request url
	resp, err := http.Get(r.RequestUrl)
	if err != nil {
		return newBlankAPIGatewayProxyResponse(), err
	}

	// create xml from example service request body
	reqXml := etree.NewDocument()
	_, err = reqXml.ReadFrom(resp.Body)

	updateRequestXml(r.RequestMap, *reqXml)

	// write it to a string for request body
	s, err := reqXml.WriteToString()
	if err != nil {
		return newBlankAPIGatewayProxyResponse(), err
	}

	// create the service request and set the request headers
	request, err := http.NewRequest(r.RequestMethod, r.ServiceUrl, strings.NewReader(s))
	if err != nil {
		return newBlankAPIGatewayProxyResponse(), err
	}
	for k, v := range r.RequestProperties {
		request.Header.Add(k, v)
	}

	// create client and get response from service
	client := &http.Client{}
	resp, err = client.Do(request)
	if err != nil {
		return newBlankAPIGatewayProxyResponse(), err
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return newBlankAPIGatewayProxyResponse(), err
		}
	default:
		reader = resp.Body
	}

	doc, err := xmlquery.Parse(reader)
	if err != nil {
		return newBlankAPIGatewayProxyResponse(), err
	}

	xml, err := xmlquery.Parse(strings.NewReader(doc.InnerText()))
	if err != nil {
		return newBlankAPIGatewayProxyResponse(), err
	}

	var results []M
	for kString, vMap := range r.ResponseMap {
		for _, e := range xmlquery.Find(xml, "//"+kString) {
			m := make(map[string]interface{})
			for k := range vMap {
				m[k] = xmlquery.FindOne(e, "//"+k).InnerText()
			}
			results = append(results, m)
		}
	}

	// marshal map into json string for response body
	j, _ := json.Marshal(results)
	return newAPIGatewayProxyResponse(string(j)), nil
}

func newBlankAPIGatewayProxyResponse() (r events.APIGatewayProxyResponse) {
	return newAPIGatewayProxyResponse("")
}

func newAPIGatewayProxyResponse(body string) (r events.APIGatewayProxyResponse) {
	return events.APIGatewayProxyResponse{
		Body:            body,
		StatusCode:      200,
		IsBase64Encoded: false,
	}
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
