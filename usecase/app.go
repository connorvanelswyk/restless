package usecase

import (
	"../model"
	"../repo"
	"compress/gzip"
	"encoding/json"
	"github.com/antchfx/xmlquery"
	"github.com/aws/aws-lambda-go/events"
	"io"
	"log"
	"net/http"
	"strings"
)

// M is an alias for map[string]interface{}
type M map[string]interface{}

func Handle(in events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sr, err := model.NewSoaplessRequest(in)
	if err != nil {
		return Error(err)
	}

	resp, err := repo.NewSoapResponse(*sr)
	if err != nil {
		return Error(err)
	}

	j, err := NewJsonResponseBody(*resp, *sr)
	if err != nil {
		return Error(err)
	}

	return Success(j)
}

func NewJsonResponseBody(r http.Response, sr model.SoaplessRequest) (string, error) {
	var reader io.ReadCloser
	var err error
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(r.Body)
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
