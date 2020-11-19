avs:
	go install ./...

run: avs
	avs

test:
	go test ./...
