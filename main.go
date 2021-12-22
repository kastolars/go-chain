package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"time"
)

const (
	TIMESTAMP_BYTES_LEN  = 8
	PREV_HASH_BYTES_LEN  = 32
	DATA_HASH_BYTES_LEN  = 32
	DIFFICULTY_BYTES_LEN = 16
	NONCE_BYTES_LEN      = 8
)

type BlockHeader struct {
	Timestamp    int64
	PreviousHash [32]byte
	DataHash     [32]byte
	Difficulty   uint16
	Nonce        uint64
}

type Block struct {
	BlockHeader
	Data []byte
}

func hashData(buf []byte) [32]byte {
	return sha256.Sum256(buf)
}

func (b BlockHeader) serialize() []byte {
	bufLen := TIMESTAMP_BYTES_LEN + PREV_HASH_BYTES_LEN + DATA_HASH_BYTES_LEN + DIFFICULTY_BYTES_LEN + NONCE_BYTES_LEN
	buf := make([]byte, bufLen)
	binary.BigEndian.PutUint64(buf[:TIMESTAMP_BYTES_LEN], uint64(b.Timestamp))
	currPos := TIMESTAMP_BYTES_LEN
	currPos += copy(buf[currPos:currPos+PREV_HASH_BYTES_LEN], b.PreviousHash[:])
	currPos += copy(buf[currPos:currPos+DATA_HASH_BYTES_LEN], b.DataHash[:])
	binary.BigEndian.PutUint16(buf[currPos:currPos+DIFFICULTY_BYTES_LEN], b.Difficulty)
	currPos += DIFFICULTY_BYTES_LEN
	binary.BigEndian.PutUint64(buf[currPos:], b.Nonce)

	return buf
}

func deserialize(buf []byte) Block {
	timestamp := int64(binary.BigEndian.Uint64(buf[:TIMESTAMP_BYTES_LEN]))
	currPos := TIMESTAMP_BYTES_LEN

	previousHash := new([32]byte)
	currPos += copy(previousHash[:], buf[currPos:currPos+PREV_HASH_BYTES_LEN])

	dataHash := new([32]byte)
	currPos += copy(dataHash[:], buf[currPos:currPos+DATA_HASH_BYTES_LEN])

	difficulty := binary.BigEndian.Uint16(buf[currPos : currPos+DIFFICULTY_BYTES_LEN])
	currPos += DIFFICULTY_BYTES_LEN

	nonce := binary.BigEndian.Uint64(buf[currPos : currPos+NONCE_BYTES_LEN])
	currPos += NONCE_BYTES_LEN

	data := buf[currPos:]

	return Block{
		BlockHeader: BlockHeader{
			Timestamp:    timestamp,
			PreviousHash: *previousHash,
			DataHash:     *dataHash,
			Difficulty:   difficulty,
			Nonce:        nonce,
		},
		Data: data,
	}

}

func (b BlockHeader) hash() [32]byte {
	buf := b.serialize()
	return hashData(buf)
}

func newBlock(previousHash [32]byte, data []byte, difficulty uint16, nonce uint64) Block {
	dataHash := hashData(data)
	return Block{
		BlockHeader: BlockHeader{
			Timestamp:    time.Now().Unix(),
			PreviousHash: previousHash,
			DataHash:     dataHash,
			Difficulty:   difficulty,
			Nonce:        nonce,
		},
		Data: data,
	}
}

func validate(hash [32]byte, difficulty *big.Int) int {
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
			Difficulty:   2,
			Nonce:        0,
		},
		Data: someRandomData,
	}

}

func start() {
	// Create genesis Block
	someRandomData := []byte("Let's Go!")
	genesisBlock := createGenesisBlock(someRandomData)

	// Create the difficulty
	difficulty := setDifficulty(int(genesisBlock.Difficulty)) // [0, 32)

	// Initialize the  blockchain in memory
	blockchain := []Block{genesisBlock}

	// Start mining
	previousHash := genesisBlock.hash()
	for {
		fmt.Printf("Blockchain length: %d\n", len(blockchain))
		var nonce uint64 = 0
		for {
			block := newBlock(previousHash, someRandomData, genesisBlock.Difficulty, nonce)
			blockHash := block.hash()
			valueDifference := validate(blockHash, difficulty)
			if valueDifference < 0 {
				blockchain = append(blockchain, block)
				previousHash = blockHash
				break
			}
			nonce++
		}
	}
}

func main() {
	start()
}
