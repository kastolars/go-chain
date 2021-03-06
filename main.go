package main

import (
	"encoding/hex"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"go-chain/block"
	"go-chain/p2p"
	"go-chain/util"
)

func run() error {
	// Post
	port := os.Args[1]

	// Collection of peers
	peers := make([]p2p.Peer, 0)

	// Thread synchronization
	peersLock := sync.Mutex{}
	blockChannel := make(chan block.Block)
	chainSyncChannel := make(chan net.Conn)

	// Create a peer id
	myPeerId := make([]byte, 32)
	rand.Seed(time.Now().UnixNano())
	rand.Read(myPeerId)
	util.GoChainLogger.Println("My peer id is: " + hex.EncodeToString(myPeerId))

	// Reads from peer
	peerHandler := p2p.HandlePeer

	if len(os.Args) > 2 {
		// First connect
		peerAddr := os.Args[2]
		if conn, err := net.Dial("tcp", peerAddr); err == nil {
			util.GoChainLogger.Println("Successfully connected to " + peerAddr)
			p := p2p.Peer{
				C:                conn,
				BlockChannel:     blockChannel,
				ChainSyncChannel: chainSyncChannel,
			}
			peers = append(peers, p)

			// Sync chains
			if err := p2p.SendChainSyncRequest(conn); err != nil {
				conn.Close()
			} else {
				// TODO: Chain sync must be synchronous
				go peerHandler(p, peers, &peersLock)
			}
		} else {
			util.GoChainLogger.Println("Unable to connect to peer: ", peerAddr)
		}
	}

	// Then listen
	listener, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		return err
	}
	defer listener.Close()

	util.GoChainLogger.Println("Listening on port 8331...")

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			util.GoChainLogger.Printf("Peer %s connected", conn.RemoteAddr().String())
			// TODO: Should complete a handshake here
			p := p2p.Peer{
				C:                conn,
				BlockChannel:     blockChannel,
				ChainSyncChannel: chainSyncChannel,
			}
			peers = append(peers, p)
			go peerHandler(p, peers, &peersLock)
		}
	}()

	// Random data we're going to use
	someRandomData := []byte("Let's Go!")

	// Create the difficultyBigInt
	var bitshift uint8 = 17
	difficultyBigInt := util.CalculateDifficulty(bitshift) // [0, 32)

	// Initialize the  blockchain in memory
	blockchain := []block.Block{}

	// Start mining
	util.GoChainLogger.Println("Mining...")
	previousHash := *new([32]byte)
	var nonce uint64 = 0
	start := time.Now().UnixNano()

	for {
		select {
		case unsyncdPeer := <-chainSyncChannel:
			for _, block := range blockchain {
				if err := p2p.SendBlock(unsyncdPeer, block); err != nil {
					util.GoChainLogger.Println("Failed to send block to peer: " + err.Error())
				}
			}
		case candidateBlock := <-blockChannel:
			blockHash := candidateBlock.Header.Hash()
			// TODO: Validate that block is good
			if !block.ValidateBlock(candidateBlock, previousHash, bitshift) {
				break
			}

			// Append to blockchain
			blockchain = append(blockchain, candidateBlock)

			// Reset dynamic values
			previousHash = blockHash
			nonce = 0
			start = time.Now().UnixNano()

			util.GoChainLogger.Printf("Block #%d, %s received", len(blockchain), hex.EncodeToString(blockHash[:]))

		default:

			// Attempt to mine a newBlock here
			newBlock := block.NewBlock(previousHash, someRandomData, bitshift, nonce)
			blockHash := newBlock.Header.Hash()
			valueDifference := util.CompareBigInt(blockHash, difficultyBigInt)

			if valueDifference >= 0 {
				// Retry
				nonce++
			} else {

				// Append to blockchain
				blockchain = append(blockchain, newBlock)

				// Metrics
				end := time.Now().UnixNano()
				elapsed := float64(end - start)

				// Reset dynamic values
				previousHash = blockHash
				nonce = 0
				start = end

				// Distribute to peers
				for _, peer := range peers {
					if err := p2p.SendBlock(peer.C, newBlock); err != nil {
						util.GoChainLogger.Println("Failed to send block to peer: " + err.Error())
					}
				}

				util.GoChainLogger.Printf("Block #%d, %s mined in %f seconds", len(blockchain), hex.EncodeToString(blockHash[:]), elapsed/1000000000.0)
			}
		}
	}
}

func main() {
	run()
}
