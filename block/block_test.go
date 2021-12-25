package block

import (
	"go-chain/util"
	"testing"
)

func TestSerialization(t *testing.T) {
	previousHash := *new([32]byte)
	data := []byte("test data")
	block := NewBlock(previousHash, data, 5, 0)
	serializedBlock := block.Serialize()
	deserializedBlock := Deserialize(serializedBlock, data)
	if deserializedBlock.Header.Timestamp != block.Header.Timestamp {
		t.Errorf("Timestamps do not match; wanted %d got %d", block.Header.Timestamp, deserializedBlock.Header.Timestamp)
	}
	if !util.CompareSize32ByteSlices(deserializedBlock.Header.PreviousHash, block.Header.PreviousHash) {
		t.Errorf("Previous hashes do not match; wanted %d got %d", block.Header.PreviousHash, deserializedBlock.Header.PreviousHash)
	}
	if !util.CompareSize32ByteSlices(deserializedBlock.Header.DataHash, block.Header.DataHash) {
		t.Errorf("Data hashes do not match; wanted %d got %d", block.Header.DataHash, deserializedBlock.Header.DataHash)
	}
	if deserializedBlock.Header.BitShift != block.Header.BitShift {
		t.Errorf("Bitshifts do not match; wanted %d got %d", block.Header.BitShift, deserializedBlock.Header.BitShift)
	}
	if deserializedBlock.Header.Nonce != block.Header.Nonce {
		t.Errorf("Nonces do not match; wanted %d got %d", block.Header.Nonce, deserializedBlock.Header.Nonce)
	}
}

func TestValidate(t *testing.T) {
	previousHash := *new([32]byte)
	data := []byte("test data")
	var bitshift uint8 = 0
	difficultyBigInt := util.CalculateDifficulty(bitshift)

	block := NewBlock(previousHash, data, bitshift, 0)
	blockHash := block.Header.Hash()
	for util.CompareBigInt(blockHash, difficultyBigInt) >= 0 {
		block.Header.Nonce++
		blockHash = block.Header.Hash()
	}
	if !ValidateBlock(block, previousHash, bitshift) {
		t.Errorf("Expected valid block")
	}
}
