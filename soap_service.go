package main

import (
	"github.com/beevik/etree"
	"golang.org/x/text/encoding/charmap"
	"io"
	"net/http"
	"strings"
)

func NewSoapResponse(sr SoaplessRequest) (*http.Response, error) {
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

	client := &http.Client{}
	response, err := client.Do(request)

	return response, err
}

func CharsetReader(charset string, input io.Reader) (io.Reader, error) {
	return charmap.ISO8859_1.NewDecoder().Reader(input), nil
}
