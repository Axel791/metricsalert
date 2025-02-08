all:
	go build -o cmd/agent/agent ./cmd/agent
	go build -o cmd/server/server ./cmd/server