.PHONY: compile run ship

compile:
	GOOS=linux GOARCH=amd64 go build -o soapless

run: compile
	sam local invoke "soapless" -e ./usecase/testdata/event-1.json | jq
	sam local invoke "soapless" -e ./usecase/testdata/event-2.json | jq
	sam local invoke "soapless" -e ./usecase/testdata/event-3.json | jq

ship: compile
	zip soapless_lambda.zip soapless