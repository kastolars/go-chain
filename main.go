package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"time"
)

const (
	TIMESTAMP_BYTES_LEN = 8
	PREV_HASH_BYTES_LEN = 32
	DATA_HASH_BYTES_LEN = 32
	NONCE_LEN           = 8
)

type BlockHeader struct {
	Timestamp    int64
	PreviousHash [32]byte
	DataHash     [32]byte
	Nonce        uint64
}

type Block struct {
	BlockHeader
	Data []byte
}

func hashData(buf []byte) [32]byte {
	return sha256.Sum256(buf)
}

func (b BlockHeader) Hash() [32]byte {
	bufLen := TIMESTAMP_BYTES_LEN + PREV_HASH_BYTES_LEN + DATA_HASH_BYTES_LEN + NONCE_LEN
	buf := make([]byte, bufLen)
	currPos := 0
	binary.BigEndian.PutUint64(buf[:TIMESTAMP_BYTES_LEN], uint64(b.Timestamp))
	currPos += TIMESTAMP_BYTES_LEN
	currPos += copy(buf[currPos:currPos+PREV_HASH_BYTES_LEN], b.PreviousHash[:])
	currPos += copy(buf[currPos:currPos+DATA_HASH_BYTES_LEN], b.DataHash[:])
	binary.BigEndian.PutUint64(buf[currPos:], uint64(b.Nonce))

	return hashData(buf)
}

func NewBlock(previousHash [32]byte, data []byte, nonce uint64) Block {
	dataHash := hashData(data)
	return Block{
		BlockHeader: BlockHeader{
			Timestamp:    time.Now().Unix(),
			PreviousHash: previousHash,
			DataHash:     dataHash,
			Nonce:        nonce,
		},
		Data: data,
	}
}

func Validate(hash [32]byte, difficulty *big.Int) int {
	z := new(big.Int)
	z.SetBytes(hash[:])
	return z.Cmp(difficulty)
}

func setDifficulty(difficultyIndex int) *big.Int {
	// Set difficulty
	difficulty := new(big.Int)
	difficultyBytes := new([32]byte)
	difficultyBytes[difficultyIndex] = 1
	difficulty.SetBytes((*difficultyBytes)[:])
	return difficulty
}

func createGenesisBlock(someRandomData []byte) Block {
	genesisDataHash := hashData(someRandomData)
	return Block{
		BlockHeader: BlockHeader{
			Timestamp:    time.Now().Unix(),
			PreviousHash: *new([32]byte),
			DataHash:     genesisDataHash,
			Nonce:        0,
		},
		Data: someRandomData,
	}

}

func start() {
	// Create the difficulty
	difficulty := setDifficulty(2) // [0, 32)

	// Create genesis Block
	someRandomData := []byte("Let's Go!")
	genesisBlock := createGenesisBlock(someRandomData)

	// Initialize the memory blockchain
	blockchain := []Block{genesisBlock}

	// Start mining
	previousHash := genesisBlock.Hash()

	for {
		fmt.Printf("Blockchain length: %d\n", len(blockchain))
		var nonce uint64 = 0
		for {
			newBlock := NewBlock(previousHash, someRandomData, nonce)
			newBlockHash := newBlock.Hash()
			valueDifference := Validate(newBlockHash, difficulty)
			if valueDifference < 0 {
				blockchain = append(blockchain, newBlock)
				previousHash = newBlockHash
				break
			}
			nonce++
		}
	}
}

func main() {
	start()
}
