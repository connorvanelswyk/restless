package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/antchfx/xmlquery"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/beevik/etree"
	"golang.org/x/text/encoding/charmap"
	"io"
	"net/http"
	"strings"
)

// M is an alias for map[string]interface{}
type M map[string]interface{}

func Handler(ctx context.Context, in events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	r, err := NewRequester([]byte(in.Body))
	if err != nil {
		return newBlankAPIGatewayProxyResponse(), err
	}

	s, err := NewSoapRequestBody(r)
	if err != nil {
		return newBlankAPIGatewayProxyResponse(), err
	}

	request, err := http.NewRequest(r.RequestMethod, r.ServiceUrl, strings.NewReader(s))
	if err != nil {
		return newBlankAPIGatewayProxyResponse(), err
	}
	for k, v := range r.RequestProperties {
		request.Header.Add(k, v)
	}

	// create client and get response from service
	client := &http.Client{}
	resp, err := client.Do(request)
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
			if len(vMap) > 0 {
				for k := range vMap {
					m[k] = xmlquery.FindOne(e, "//"+k).InnerText()
				}
			} else {
				m[kString] = e.InnerText()
			}
			results = append(results, m)
		}
	}

	// marshal map into json string for response body
	j, _ := json.Marshal(results)
	return newAPIGatewayProxyResponse(string(j)), nil
}

func NewSoapRequestBody(r *Requester) (string, error) {
	resp, err := http.Get(r.RequestUrl)
	if err != nil {
		return "", err
	}
	doc := etree.Document{}
	if r.Encoding == "ISO-8859-1" {
		doc.ReadSettings = etree.ReadSettings{
			CharsetReader: CharsetReader,
		}
	}
	_, err = doc.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}
	if r.RequestMap != nil {
		for k, v := range r.RequestMap {
			e := doc.FindElement("//" + k)
			e.SetText(v)
		}
	}
	return doc.WriteToString()
}

func CharsetReader(charset string, input io.Reader) (io.Reader, error) {
	return charmap.ISO8859_1.NewDecoder().Reader(input), nil
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

func main() {
	lambda.Start(Handler)
}
