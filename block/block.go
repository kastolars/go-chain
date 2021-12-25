package block

import (
	"encoding/binary"
	"math/big"
	"time"

	"go-chain/util"
)

const (
	TIMESTAMP_BYTES_LEN = 8
	PREV_HASH_BYTES_LEN = 32
	DATA_HASH_BYTES_LEN = 32
	BITSHIFT_BYTES_LEN  = 1
	NONCE_BYTES_LEN     = 8
)

const BLOCK_HEADER_LEN = TIMESTAMP_BYTES_LEN + PREV_HASH_BYTES_LEN + DATA_HASH_BYTES_LEN + BITSHIFT_BYTES_LEN + NONCE_BYTES_LEN

type BlockHeader struct {
	Timestamp    int64
	PreviousHash [32]byte
	DataHash     [32]byte
	BitShift     uint8
	Nonce        uint64
}

type Block struct {
	Header BlockHeader
	Data   []byte
}

func (b BlockHeader) Serialize() []byte {
	buf := make([]byte, BLOCK_HEADER_LEN)

	// Timestamp
	binary.BigEndian.PutUint64(buf[:TIMESTAMP_BYTES_LEN], uint64(b.Timestamp))
	currPos := TIMESTAMP_BYTES_LEN

	// Previous Hash
	currPos += copy(buf[currPos:currPos+PREV_HASH_BYTES_LEN], b.PreviousHash[:])

	// Data Hash
	currPos += copy(buf[currPos:currPos+DATA_HASH_BYTES_LEN], b.DataHash[:])

	// Bit Shift
	buf[currPos] = b.BitShift
	currPos += BITSHIFT_BYTES_LEN

	// Nonce
	binary.BigEndian.PutUint64(buf[currPos:], b.Nonce)

	return buf
}

func (b Block) Serialize() []byte {
	serializedBlockHeader := b.Header.Serialize()
	dataLen := len(b.Data)
	dataLenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(dataLenBuf, uint16(dataLen))
	serializedBlock := append(serializedBlockHeader, dataLenBuf...)
	return append(serializedBlock, b.Data...)
}

func Deserialize(blockHeaderBuf []byte, dataBuf []byte) Block {
	// Timestamp
	timestamp := int64(binary.BigEndian.Uint64(blockHeaderBuf[:TIMESTAMP_BYTES_LEN]))
	currPos := TIMESTAMP_BYTES_LEN

	// Previous hash
	previousHash := new([32]byte)
	currPos += copy(previousHash[:], blockHeaderBuf[currPos:currPos+PREV_HASH_BYTES_LEN])

	// Data hash
	dataHash := new([32]byte)
	currPos += copy(dataHash[:], blockHeaderBuf[currPos:currPos+DATA_HASH_BYTES_LEN])

	// Bitshift
	bitshift := uint8(blockHeaderBuf[currPos])
	currPos += BITSHIFT_BYTES_LEN

	// Nonce
	nonce := binary.BigEndian.Uint64(blockHeaderBuf[currPos : currPos+NONCE_BYTES_LEN])
	currPos += NONCE_BYTES_LEN

	return Block{
		Header: BlockHeader{
			Timestamp:    timestamp,
			PreviousHash: *previousHash,
			DataHash:     *dataHash,
			BitShift:     bitshift,
			Nonce:        nonce,
		},
		Data: dataBuf,
	}

}

func (b BlockHeader) Hash() [32]byte {
	buf := b.Serialize()
	return util.HashData(buf)
}

func NewBlock(previousHash [32]byte, data []byte, bitshift uint8, nonce uint64) Block {
	dataHash := util.HashData(data)
	return Block{
		Header: BlockHeader{
			Timestamp:    time.Now().Unix(),
			PreviousHash: previousHash,
			DataHash:     dataHash,
			BitShift:     bitshift,
			Nonce:        nonce,
		},
		Data: data,
	}
}

func ValidateBlock(candidate Block, previousHash [32]byte, bitshift uint8, difficultyBigInt *big.Int) bool {
	// Timestamp
	if candidate.Header.Timestamp >= time.Now().UnixNano() {
		return false
	}

	// Previous hash
	if candidate.Header.PreviousHash != previousHash {
		return false
	}

	// Data hash
	dataHash := util.HashData(candidate.Data)
	if candidate.Header.DataHash != dataHash {
		return false
	}

	// Bit shift
	if candidate.Header.BitShift != bitshift {
		return false
	}

	// Compare hash to difficulty
	blockHash := candidate.Header.Hash()
	valueDifference := util.CompareBigInt(blockHash, difficultyBigInt)
	if valueDifference >= 0 {
		return false
	}

	return true
}
