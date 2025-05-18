.PHONY: build run clean init deps bootstrap node1 node2 node3 run-demo kill-demo



build:
	go build -o app main.go

run: build
	./app

bootstrap: build
	./app

join: build
	./app -seed $(SEED)

# Blacklist specific node IDs (hex encoded)
# Example: make blacklist NODE_IDS="abcd1234,5678efgh"
blacklist: build
	./app -blacklist $(NODE_IDS)



# Run a demo network with 3 nodes using fixed IDs
node1: build
	@echo "Starting bootstrap node (node1) on port 30001..."
	./app -id "node1" -port 30001

node2: build
	@echo "Starting node2 on port 30002 (connects to node1, blocks node3)..."
	./app  -port 30002 -seed  tcp://:30001 -blacklist "36653666363436353333" -id "36653666363436353332"

node3: build
	@echo "Starting node3 on port 30003 (connects to node1, blocks node2)..."
	./app  -port 30003 -seed  tcp://:30001 -blacklist "36653666363436353332" -id "36653666363436353333"

# Kill all running demo nodes
kill-demo:
	@echo "Stopping all nodes..."
	@pkill -f "./app -id" || true

clean:
	rm -f app
	rm -f go.mod go.sum
	rm -f node*.log 