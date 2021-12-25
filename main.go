package main

import (
	"encoding/hex"
	"log"
	"net"
	"os"
	"time"

	"go-chain/block"
	"go-chain/util"
)

var logger *log.Logger = log.New(os.Stdout, "LOG: ", log.Lmicroseconds|log.Lshortfile)

func p2p() error {
	port := os.Args[1]
	peers := make([]net.Conn, 0)
	newBlockChannel := make(chan block.Block)

	peerHandler := func(c net.Conn) {
		for {
			buf := make([]byte, 512)
			if _, err := c.Read(buf); err != nil {
				logger.Println("Failed to read from peer: " + err.Error())
				c.Close()
				return
			}
			block := block.Deserialize(buf)
			newBlockChannel <- block
		}
	}

	if len(os.Args) > 2 {
		// First connect
		peerAddr := os.Args[2]
		if conn, err := net.Dial("tcp", peerAddr); err == nil {
			logger.Println("Successfully connected to " + peerAddr)
			peers = append(peers, conn)
			go peerHandler(conn)
		} else {
			logger.Println("Unable to connect to peer: ", peerAddr)
		}
	}

	// Then listen
	listener, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		return err
	}
	defer listener.Close()

	logger.Println("Listening on port 8331...")

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			logger.Printf("Peer %s connected", conn.RemoteAddr().String())
			peers = append(peers, conn)
			go peerHandler(conn)
		}
	}()

	// Random data we're going to use
	someRandomData := []byte("Let's Go!")

	// Create the difficultyBigInt
	var bitshift uint8 = 16
	difficultyBigInt := util.CalculateDifficulty(bitshift) // [0, 32)

	// Initialize the  blockchain in memory
	blockchain := []block.Block{}

	// Start mining
	logger.Println("Mining...")
	previousHash := *new([32]byte)
	var nonce uint64 = 0
	start := time.Now().UnixNano()
	for {
		select {
		case candidateBlock := <-newBlockChannel:
			blockHash := candidateBlock.Header.Hash()
			// TODO: Validate that block is good

			// Append to blockchain
			blockchain = append(blockchain, candidateBlock)

			// Reset dynamic values
			previousHash = blockHash
			nonce = 0
			start = time.Now().UnixNano()

			logger.Printf("Block #%d, %s received", len(blockchain), hex.EncodeToString(blockHash[:]))

		default:
			block := block.NewBlock(previousHash, someRandomData, bitshift, nonce)
			blockHash := block.Header.Hash()
			valueDifference := util.CompareBigInt(blockHash, difficultyBigInt)
			if valueDifference < 0 {

				// Append to blockchain
				blockchain = append(blockchain, block)

				// Metrics
				end := time.Now().UnixNano()
				elapsed := float64(end - start)

				// Reset dynamic values
				previousHash = blockHash
				nonce = 0
				start = end

				// Distribute to peers
				for _, peer := range peers {
					blockBuf := block.Header.Serialize()
					if _, err := peer.Write(blockBuf); err != nil {
						logger.Println("Failed to send block to peer: " + err.Error())
					}
				}

				logger.Printf("Block #%d, %s mined in %f seconds", len(blockchain), hex.EncodeToString(blockHash[:]), elapsed/1000000000.0)
			} else {
				nonce++
			}
		}
	}
}

func main() {
	p2p()
}
