package main

import (
	"crypto/sha256"
	"encoding/binary"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

const (
	TIMESTAMP_BYTES_LEN = 8
	PREV_HASH_BYTES_LEN = 32
	DATA_HASH_BYTES_LEN = 32
	BITSHIFT_BYTES_LEN  = 1
	NONCE_BYTES_LEN     = 8
)

var logger *log.Logger = log.New(os.Stdout, "LOG: ", log.Lmicroseconds|log.Lshortfile)

type blockHeader struct {
	Timestamp    int64
	PreviousHash [32]byte
	DataHash     [32]byte
	BitShift     uint8
	Nonce        uint64
}

type block struct {
	blockHeader
	Data []byte
}

func hashData(buf []byte) [32]byte {
	return sha256.Sum256(buf)
}

func (b blockHeader) serialize() []byte {
	bufLen := TIMESTAMP_BYTES_LEN + PREV_HASH_BYTES_LEN + DATA_HASH_BYTES_LEN + BITSHIFT_BYTES_LEN + NONCE_BYTES_LEN
	buf := make([]byte, bufLen)

	// Timestamp
	binary.BigEndian.PutUint64(buf[:TIMESTAMP_BYTES_LEN], uint64(b.Timestamp))
	currPos := TIMESTAMP_BYTES_LEN

	// Previous Hash
	currPos += copy(buf[currPos:currPos+PREV_HASH_BYTES_LEN], b.PreviousHash[:])

	// Data Hash
	currPos += copy(buf[currPos:currPos+DATA_HASH_BYTES_LEN], b.DataHash[:])

	// Bit Shift
	buf[currPos+BITSHIFT_BYTES_LEN] = b.BitShift
	currPos += BITSHIFT_BYTES_LEN

	// Nonce
	binary.BigEndian.PutUint64(buf[currPos:], b.Nonce)

	return buf
}

func deserialize(buf []byte) block {
	// Timestamp
	timestamp := int64(binary.BigEndian.Uint64(buf[:TIMESTAMP_BYTES_LEN]))
	currPos := TIMESTAMP_BYTES_LEN

	// Previous hash
	previousHash := new([32]byte)
	currPos += copy(previousHash[:], buf[currPos:currPos+PREV_HASH_BYTES_LEN])

	// Data hash
	dataHash := new([32]byte)
	currPos += copy(dataHash[:], buf[currPos:currPos+DATA_HASH_BYTES_LEN])

	// Bitshift
	bitshift := uint8(buf[currPos+BITSHIFT_BYTES_LEN])
	currPos += BITSHIFT_BYTES_LEN

	// Nonce
	nonce := binary.BigEndian.Uint64(buf[currPos : currPos+NONCE_BYTES_LEN])
	currPos += NONCE_BYTES_LEN

	// Data
	data := buf[currPos:]

	return block{
		blockHeader: blockHeader{
			Timestamp:    timestamp,
			PreviousHash: *previousHash,
			DataHash:     *dataHash,
			BitShift:     bitshift,
			Nonce:        nonce,
		},
		Data: data,
	}

}

func (b blockHeader) hash() [32]byte {
	buf := b.serialize()
	return hashData(buf)
}

func newBlock(previousHash [32]byte, data []byte, bitshift uint8, nonce uint64) block {
	dataHash := hashData(data)
	return block{
		blockHeader: blockHeader{
			Timestamp:    time.Now().Unix(),
			PreviousHash: previousHash,
			DataHash:     dataHash,
			BitShift:     bitshift,
			Nonce:        nonce,
		},
		Data: data,
	}
}

func compareBigInt(hash [32]byte, difficulty *big.Int) int {
	z := new(big.Int)
	z.SetBytes(hash[:])
	return z.Cmp(difficulty)
}

func setDifficulty(difficultyIndex int) *big.Int {
	difficulty := new(big.Int)
	difficultyBytes := new([32]byte)
	difficultyBytes[difficultyIndex] = 1
	difficulty.SetBytes((*difficultyBytes)[:])
	return difficulty
}

func calculateDifficulty(bitShift uint8) *big.Int {
	difficultyBytes := new([32]byte)
	difficultyBytes[0] = 1

	difficulty := new(big.Int)
	difficulty.SetBytes((*difficultyBytes)[:])
	difficulty.Rsh(difficulty, uint(bitShift))

	return difficulty
}

func p2p() error {
	port := os.Args[1]
	peers := make([]net.Conn, 0)

	peerHandler := func(c net.Conn) {
		for {
			buf := make([]byte, 512)
			if _, err := c.Read(buf); err != nil {
				logger.Println("Failed to read from peer: " + err.Error())
				c.Close()
				return
			}
			logger.Println("Received block from peer")
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

	// Create genesis Block
	someRandomData := []byte("Let's Go!")

	// Create the difficultyBigInt
	var bitshift uint8 = 16
	difficultyBigInt := calculateDifficulty(bitshift) // [0, 32)

	// Initialize the  blockchain in memory
	blockchain := []block{}

	// Start mining
	logger.Println("Mining...")
	previousHash := *new([32]byte)
	var nonce uint64 = 0
	start := time.Now().UnixNano()
	for {
		select {
		default:
			block := newBlock(previousHash, someRandomData, bitshift, nonce)
			blockHash := block.hash()
			valueDifference := compareBigInt(blockHash, difficultyBigInt)
			if valueDifference < 0 {
				blockchain = append(blockchain, block)
				previousHash = blockHash
				end := time.Now().UnixNano()
				elapsed := float64(end - start)
				start = end
				logger.Printf("Block #%d took: %f seconds", len(blockchain), elapsed/1000000000.0)

				// Distribute to peers
				for _, peer := range peers {
					blockBuf := block.serialize()
					if _, err := peer.Write(blockBuf); err != nil {
						logger.Println("Failed to send block to peer: " + err.Error())
					}
					logger.Println("Sent block to peer")
				}
			} else {
				nonce++
			}
		}
	}
}

func main() {
	p2p()
}
