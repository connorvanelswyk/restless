.PHONY:compile
.PHONY:run
.PHONY:ship

compile:
	GOOS=linux GOARCH=amd64 go build -o restless

run: compile
	sam local invoke "restless" -e ./events/event-1.json | jq

ship: compile
	zip restless_lambda.zip restless