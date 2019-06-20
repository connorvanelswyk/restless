package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/antchfx/xmlquery"
	"github.com/aws/aws-lambda-go/events"
	"github.com/beevik/etree"
	"golang.org/x/text/encoding/charmap"
	"io"
	"log"
	"net/http"
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

// M is an alias for map[string]interface{}
type M map[string]interface{}

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

func NewHttpRequest(sr SoaplessRequest) (*http.Request, error) {
	resp, err := http.Get(sr.RequestBody)
	if err != nil {
		return nil, err
	}
	doc := etree.Document{}
	if sr.Encoding == "ISO-8859-1" {
		doc.ReadSettings = etree.ReadSettings{
			CharsetReader: CharsetReader,
		}
	}
	_, err = doc.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	if sr.RequestMap != nil {
		for k, v := range sr.RequestMap {
			e := doc.FindElement("//" + k)
			e.SetText(v)
		}
	}
	s, err := doc.WriteToString()
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(sr.RequestMethod, sr.Service, strings.NewReader(s))
	if err != nil {
		return nil, err
	}
	for k, v := range sr.RequestProperties {
		request.Header.Add(k, v)
	}

	return request, nil
}

func CharsetReader(charset string, input io.Reader) (io.Reader, error) {
	return charmap.ISO8859_1.NewDecoder().Reader(input), nil
}

func NewJsonResponseBody(r http.Response, sr SoaplessRequest) (string, error) {
	var reader io.ReadCloser
	var err error
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(r.Body)
		break
	default:
		reader = r.Body
	}
	if err != nil {
		return "", err
	}

	doc, err := xmlquery.Parse(reader)
	if err != nil {
		return "", err
	}

	xml, err := xmlquery.Parse(strings.NewReader(doc.InnerText()))
	if err != nil {
		return "", err
	}

	var results []M
	for kString, vMap := range sr.ResponseMap {
		for _, node := range doc.SelectElements(kString) {
			log.Println(node.InnerText())
		}
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
	j, err := json.Marshal(results)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

func Error(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:            err.Error(),
		StatusCode:      400,
		IsBase64Encoded: false,
	}, err
}

func Success(json string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:            json,
		StatusCode:      200,
		IsBase64Encoded: false,
	}, nil
}
