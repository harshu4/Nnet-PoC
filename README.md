# nnet Node Network - Proof of Concept

This is a Proof of Concept (PoC) demonstrating how messages are routed in a P2P network when nodes blacklist each other. Built using [nknorg/nnet](https://github.com/nknorg/nnet), this project shows that even when node2 and node3 blacklist each other, they can still communicate through the seed node (node1) acting as an intermediary.

## Key Demonstration

This PoC highlights:

1. **Indirect Communication**: When node2 and node3 blacklist each other, they cannot establish direct connections, but can still exchange messages via node1 (the seed node)
2. **Message Routing**: Messages from node2 to node3 (and vice versa) are automatically routed through node1
3. **Blacklist Effectiveness**: The blacklist successfully prevents direct connections between specific nodes

## Features

- **Bootstrap Node Creation**: Set up the first node in the network (the seed node)
- **Node Discovery and Blacklisting**: Join existing networks while blacklisting specific nodes
- **Message Relay**: Send messages that are relayed through intermediary nodes
- **Interactive CLI**: Command-line interface for managing connections and sending messages

## Prerequisites

- Go 1.15+

## Installation

Clone the repository and build:

```bash
git clone https://github.com/your-username/testnnet.git
cd testnnet
make build
```

## Running the Demonstration


### Running Individual Nodes

Start each node separately to interact with them directly:

```bash
# Terminal 1
make node1  # Seed node

# Terminal 2
make node2  # Blocks node3


# Terminal 3
make node3  # Blocks node2
```

Now go to terminal 2 and type 

```bash

send 36653666363436353333 hello 

```
The message will be send to node1 instead of node3.

### CLI Commands

Once a node is running, you can interact with it using these commands:

- `help` - Show available commands
- `exit` - Exit the program
- `peers` - List connected peers
- `send <node_id> <message>` - Send message to specific node (even blacklisted ones!)
- `broadcast <message>` - Broadcast message to all connected peers
- `blacklist <node_id>` - Add a node to blacklist
- `unblacklist <node_id>` - Remove a node from blacklist

## How the PoC Works

### Node Blacklisting

The blacklisting system works at the connection level using the `WillConnectToNode` middleware:

1. When a connection attempt is made, the node ID is checked against the blacklist
2. If the ID is blacklisted, the connection is rejected before it's established
3. Direct connections between blacklisted nodes are prevented

### Message Routing

Messages are still delivered between blacklisted nodes because:

1. Both nodes maintain connections to the seed node (node1)
2. The nnet library automatically routes messages through common connections
3. When node2 sends a message to node3, it's routed through node1, which then forwards it to node3

## Project Structure

- `main.go` - Main application code
- `Makefile` - Build and run commands
- `README.md` - Project documentation

## Acknowledgments

- [nknorg/nnet](https://github.com/nknorg/nnet) - The underlying P2P network library 