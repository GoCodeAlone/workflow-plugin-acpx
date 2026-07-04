.PHONY: build test install-local clean

build:
	go build -o workflow-plugin-acpx ./cmd/workflow-plugin-acpx

test:
	go test ./...

install-local: build
	wfctl plugin install --local .

clean:
	rm -f workflow-plugin-acpx

