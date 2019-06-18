curl \
-d @cmd/event.json \
-H 'Content-Type: application/json' \
https://gfv1670v1c.execute-api.us-east-1.amazonaws.com/release | jq

curl \
-d @cmd/event-body.json \
-H 'Content-Type: application/json' \
https://gfv1670v1c.execute-api.us-east-1.amazonaws.com/release | jq

curl \
-d @cmd/event-body-2.json \
-H 'Content-Type: application/json' \
https://gfv1670v1c.execute-api.us-east-1.amazonaws.com/release | jq

curl \
-d @cmd/event-body-3.json \
-H 'Content-Type: application/json' \
https://gfv1670v1c.execute-api.us-east-1.amazonaws.com/release | jq