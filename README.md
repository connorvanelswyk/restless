[![Build Status][ci-img]][ci]
[![Coverage Status][coveralls-img]][coveralls]
[![Report Card][go-img]][go-url]

# RESTless (alpha)

A serverless, RESTful web service for painless SOAP message orchestration.

**Working examples** with Weather.gov SOAP web service:
```cmd
curl \
-d '{
  "service": "https://graphical.weather.gov/xml/SOAP_server/ndfdXMLserver.php",
  "requestBody": "https://graphical.weather.gov/xml/docs/SOAP_Requests/LatLonListZipCode.xml",
  "responseMap": { "latLonList": {} }
}' \
-H 'Content-Type: application/json' \
https://gfv1670v1c.execute-api.us-east-1.amazonaws.com/release | jq
```

```cmd
curl \
-d '{
  "serviceUrl": "https://graphical.weather.gov:443/xml/SOAP_server/ndfdXMLserver.php",
  "requestUrl": "https://graphical.weather.gov/xml/docs/SOAP_Requests/LatLonListSquare.xml",  
  "responseMap": { "latLonList": {} }
}' \
-H 'Content-Type: application/json' \
https://gfv1670v1c.execute-api.us-east-1.amazonaws.com/release | jq
```

```cmd
curl \
-d '{
  "serviceUrl": "https://graphical.weather.gov:443/xml/SOAP_server/ndfdXMLserver.php",
  "requestUrl": "https://graphical.weather.gov/xml/docs/SOAP_Requests/GmlLatLonList.xml",  
  "requestMap": {
    "requestedTime": "2019-06-22T23:59:59"
  },
  "responseMap": {
    "gml:boundedBy": {
      "gml:coordinates": ""
    },
    "gml:featureMember": {
      "gml:coordinates": "",
      "app:validTime": "",
      "app:maximumTemperature": ""
    }
  }
}' \
-H 'Content-Type: application/json' \
https://gfv1670v1c.execute-api.us-east-1.amazonaws.com/release | jq
```

[ci-img]: https://travis-ci.com/connorvanelswyk/restless.svg?branch=master
[ci]: https://travis-ci.com/connorvanelswyk/restless
[coveralls-img]: https://coveralls.io/repos/github/connorvanelswyk/restless/badge.svg?branch=master
[coveralls]: https://coveralls.io/github/connorvanelswyk/restless
[go-img]: https://goreportcard.com/badge/github.com/connorvanelswyk/restless
[go-url]: https://goreportcard.com/report/github.com/connorvanelswyk/restless