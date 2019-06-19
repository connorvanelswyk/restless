package main

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

type Requester struct {
	ServiceUrl        string                       `json:"serviceUrl"` // The endpoint for the desired SOAP service
	RequestUrl        string                       `json:"requestUrl"` // The location of an empty or sample SOAP service request
	RequestMethod     string                       `json:"requestMethod,omitempty"`
	Encoding          string                       `json:"encoding,omitempty"`
	RequestProperties map[string]string            `json:"requestProperties,omitempty"`
	RequestMap        map[string]string            `json:"requestMap,omitempty"`
	ResponseMap       map[string]map[string]string `json:"responseMap"`
}

const defaultUserAgent = "Apache-HttpClient/4.1.1"
const defaultAcceptEncoding = "gzip,deflate"

func NewRequester(inputBody []byte) (*Requester, error) {
	r := &Requester{}
	if err := json.Unmarshal(inputBody, r); err != nil {
		return r, err
	}
	if _, err := url.ParseRequestURI(r.ServiceUrl); err != nil {
		return r, errors.New("service url is malformed")
	}
	if _, err := url.ParseRequestURI(r.RequestUrl); err != nil {
		return r, errors.New("request url is malformed")
	}
	if r.Encoding == "" {
		r.Encoding = "ISO-8859-1"
	}
	if r.RequestProperties == nil {
		r.RequestProperties = map[string]string{
			"Host":            Host(r.ServiceUrl),
			"User-Agent":      defaultUserAgent,
			"Content-Type":    "text/xml;charset=" + r.Encoding,
			"Accept-Encoding": defaultAcceptEncoding,
		}
	}
	if r.RequestMethod == "" {
		r.RequestMethod = "POST"
	}
	return r, nil
}

func Host(url string) string {
	url = strings.Replace(url, "http://", "", -1)
	url = strings.Replace(url, "https://", "", -1)
	return url[:strings.Index(url, "/")]
}
