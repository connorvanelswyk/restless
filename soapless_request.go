package main

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"net/url"
	"strings"
)

type SoaplessRequest struct {
	Service           string                       `json:"service"`     // The endpoint for the desired SOAP service
	RequestBody       string                       `json:"requestBody"` // The location of an empty or sample SOAP service request
	RequestMethod     string                       `json:"requestMethod,omitempty"`
	Encoding          string                       `json:"encoding,omitempty"`
	RequestProperties map[string]string            `json:"requestProperties,omitempty"`
	RequestMap        map[string]string            `json:"requestMap,omitempty"`
	ResponseMap       map[string]map[string]string `json:"responseMap"`
}

func NewSoaplessRequest(input events.APIGatewayProxyRequest) (*SoaplessRequest, error) {
	r := &SoaplessRequest{}
	if err := json.Unmarshal([]byte(input.Body), r); err != nil {
		return nil, err
	}
	if _, err := url.ParseRequestURI(r.Service); err != nil {
		return nil, errors.New("service url is malformed")
	}
	if _, err := url.ParseRequestURI(r.RequestBody); err != nil {
		return nil, errors.New("request url is malformed")
	}
	if r.Encoding == "" {
		r.Encoding = "ISO-8859-1"
	}
	if r.RequestProperties == nil {
		host := strings.Replace(r.Service, "http://", "", -1)
		host = strings.Replace(host, "https://", "", -1)
		host = host[:strings.Index(host, "/")]
		r.RequestProperties = map[string]string{
			"Host":            host,
			"User-Agent":      "Apache-HttpClient/4.1.1",
			"Content-Type":    "text/xml;charset=" + r.Encoding,
			"Accept-Encoding": "gzip,deflate",
		}
	}
	if r.RequestMethod == "" {
		r.RequestMethod = "POST"
	}
	return r, nil
}
