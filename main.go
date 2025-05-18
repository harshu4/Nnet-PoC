// nnet Node Network - Proof of Concept
//
// This application demonstrates a peer-to-peer network using nknorg/nnet where:
// 1. Three nodes are created: node1 (bootstrap), node2, and node3
// 2. Node2 blacklists node3, and node3 blacklists node2
// 3. Node2 and node3 both connect to node1 (the bootstrap node)
// 4. Despite mutual blacklisting, node2 and node3 can still exchange messages via node1
//
// The key demonstration is that nodes can communicate indirectly through
// intermediary nodes even when they blacklist each other directly.
//
// For more information, see the README.md file.

package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/nknorg/nnet"
	"github.com/nknorg/nnet/node"
	pbmsg "github.com/nknorg/nnet/protobuf/message"
	pbnode "github.com/nknorg/nnet/protobuf/node"
)

// BlackList maintains a list of blocked node IDs
type BlackList struct {
	blockedIDs map[string]bool
	mutex      sync.RWMutex
}

// NewBlackList creates a new blacklist
func NewBlackList() *BlackList {
	return &BlackList{
		blockedIDs: make(map[string]bool),
	}
}

// AddID adds a node ID to the blacklist
func (b *BlackList) AddID(id string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.blockedIDs[id] = true
	fmt.Printf("Added node ID %s to blacklist\n", id)
}

// RemoveID removes a node ID from the blacklist
func (b *BlackList) RemoveID(id string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	delete(b.blockedIDs, id)
	fmt.Printf("Removed node ID %s from blacklist\n", id)
}

// IsBlocked checks if a node ID is blacklisted
func (b *BlackList) IsBlocked(id string) bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.blockedIDs[id]
}

func generateRandomID(length int) ([]byte, error) {
	id := make([]byte, length)
	_, err := rand.Read(id)
	if err != nil {
		return nil, err
	}
	return id, nil
}

// printHelp prints available commands
func printHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  help                          - Show this help message")
	fmt.Println("  exit                          - Exit the program")
	fmt.Println("  peers                         - List connected peers")
	fmt.Println("  send <node_id> <message>      - Send message to specific node")
	fmt.Println("  broadcast <message>           - Broadcast message to all peers")
	fmt.Println("  blacklist <node_id>           - Add a node to blacklist")
	fmt.Println("  unblacklist <node_id>         - Remove a node from blacklist")
}

func main() {
	// Parse command line flags
	bootstrapAddr := flag.String("seed", "", "Seed node address to join (empty for bootstrap node)")
	port := flag.Int("port", 30001, "Local port to listen on")
	nodeID := flag.String("id", "", "Node identifier (optional, random if not specified)")
	blacklistIDs := flag.String("blacklist", "", "Comma-separated node IDs to blacklist (hex encoded)")
	flag.Parse()

	// Initialize blacklist
	blacklist := NewBlackList()
	if *blacklistIDs != "" {
		ids := strings.Split(*blacklistIDs, ",")
		for _, id := range ids {
			blacklist.AddID(strings.TrimSpace(id))
		}
	}

	// Create node info
	var id []byte
	var err error

	// Create nnet config with custom port
	conf := &nnet.Config{
		Port:      uint16(*port),
		Transport: "tcp",
	}
	var nn *nnet.NNet
	// Create nnet instance
	if len(*nodeID) == 0 {
		nn, err = nnet.NewNNet(nil, conf)
	} else {
		id, _ = hex.DecodeString(*nodeID)
		nn, err = nnet.NewNNet(id, conf)
	}
	if err != nil {
		log.Fatalf("Create nnet error: %v", err)
	}

	// Set up middleware for connection handling using WillConnectToNode
	nn.MustApplyMiddleware(node.WillConnectToNode{func(n *pbnode.Node) (bool, bool) {
		// If node has no ID yet, allow connection to proceed
		if n == nil || len(n.Id) == 0 {
			return true, true
		}

		remoteIDHex := hex.EncodeToString(n.Id)

		// Check if the node ID is blacklisted
		if blacklist.IsBlocked(remoteIDHex) {
			fmt.Printf("Blocked connection attempt to blacklisted node ID: %s\n", remoteIDHex)
			return false, true // Don't connect, but continue middleware chain
		}

		fmt.Printf("Allowing connection to node with ID: %s\n", remoteIDHex)
		return true, true
	}, 0})

	// Also add a RemoteNodeConnected handler to log connections and show IDs
	nn.MustApplyMiddleware(node.RemoteNodeConnected{func(remoteNode *node.RemoteNode) bool {
		remoteIDHex := hex.EncodeToString(remoteNode.Id)
		fmt.Printf("Remote node connected: %s (ID: %s)\n", remoteNode.Addr, remoteIDHex)
		return true
	}, 0})

	// Set up message handler
	nn.MustApplyMiddleware(node.BytesReceived{func(msg, msgID, srcID []byte, remoteNode *node.RemoteNode) ([]byte, bool) {
		senderIDHex := hex.EncodeToString(srcID)
		fmt.Printf("\nMessage from %s: %s\n> ", senderIDHex, string(msg))

		// Auto-reply with confirmation
		nn.SendBytesRelayReply(msgID, []byte("Message received"), srcID)

		return msg, true
	}, 0})

	// --- Start Network ---
	isBootstrapNode := (*bootstrapAddr == "")

	// Start nnet
	err = nn.Start(isBootstrapNode)
	if err != nil {
		log.Fatalf("Start nnet error: %v", err)
	}

	if isBootstrapNode {
		fmt.Println("Starting as bootstrap node")
	} else {
		fmt.Printf("Joining network through seed node: %s\n", *bootstrapAddr)
		// Join the network by connecting to seed node
		err = nn.Join(*bootstrapAddr)
		if err != nil {
			log.Fatalf("Join network error: %v", err)
		}
	}

	// Print node information
	idStr := hex.EncodeToString(id)
	localAddr := nn.GetLocalNode().Addr
	localPort := nn.GetConfig().Port
	fmt.Printf("Node ID: %s\n", idStr)
	fmt.Printf("Node listening at: %s:%d\n", localAddr, localPort)

	// Command-line interface
	fmt.Println("\nNode is running. Type 'help' for available commands.")

	printHelp()

	// Start a separate goroutine to keep the node alive
	go func() {
		for {
			time.Sleep(30 * time.Second)
		}
	}()

	// Handle user input
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		input := scanner.Text()
		args := strings.Fields(input)

		if len(args) == 0 {
			fmt.Print("> ")
			continue
		}

		switch args[0] {
		case "help":
			printHelp()

		case "exit":
			fmt.Println("Exiting...")
			nn.Stop(nil)
			return

		case "peers":
			// Display connected peers
			neighbors, err := nn.GetLocalNode().GetNeighbors(nil)
			if err != nil {
				fmt.Printf("Error getting neighbors: %v\n", err)
				break
			}

			if len(neighbors) == 0 {
				fmt.Println("No peers connected")
			} else {
				fmt.Printf("Connected to %d peers:\n", len(neighbors))
				for i, peer := range neighbors {
					peerIDHex := hex.EncodeToString(peer.Id)
					fmt.Printf("  %d. %s (ID: %s)\n", i+1, peer.Addr, peerIDHex)
				}
			}

		case "send":
			if len(args) < 3 {
				fmt.Println("Usage: send <node_id> <message>")
				break
			}

			targetID, err := hex.DecodeString(args[1])
			if err != nil {
				fmt.Printf("Invalid node ID format: %v\n", err)
				break
			}

			message := strings.Join(args[2:], " ")
			reply, senderID, err := nn.SendBytesRelaySync([]byte(message), targetID)
			if err != nil {
				fmt.Printf("Failed to send message: %v\n", err)
			} else {
				senderIDHex := hex.EncodeToString(senderID)
				fmt.Printf("Message sent. Reply from %s: %s\n", senderIDHex, string(reply))
			}

		case "broadcast":
			if len(args) < 2 {
				fmt.Println("Usage: broadcast <message>")
				break
			}

			message := strings.Join(args[1:], " ")
			_, err := nn.SendBytesBroadcastAsync([]byte(message), pbmsg.RoutingType_BROADCAST_PUSH)
			if err != nil {
				fmt.Printf("Failed to broadcast message: %v\n", err)
			} else {
				fmt.Println("Message broadcast to all peers")
			}

		case "blacklist":
			if len(args) != 2 {
				fmt.Println("Usage: blacklist <node_id>")
				break
			}

			blacklist.AddID(args[1])

		case "unblacklist":
			if len(args) != 2 {
				fmt.Println("Usage: unblacklist <node_id>")
				break
			}

			blacklist.RemoveID(args[1])

		default:
			fmt.Printf("Unknown command: %s\n", args[0])
			fmt.Println("Type 'help' for available commands")
		}

		fmt.Print("> ")
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v\n", err)
	}
}
