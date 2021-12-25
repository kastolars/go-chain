package util

import (
	"crypto/sha256"
	"math/big"
)

func CalculateDifficulty(bitShift uint8) *big.Int {
	difficultyBytes := new([32]byte)
	difficultyBytes[0] = 1

	difficulty := new(big.Int)
	difficulty.SetBytes((*difficultyBytes)[:])
	difficulty.Rsh(difficulty, uint(bitShift))

	return difficulty
}

func CompareBigInt(hash [32]byte, difficulty *big.Int) int {
	z := new(big.Int)
	z.SetBytes(hash[:])
	return z.Cmp(difficulty)
}

func SetDifficulty(difficultyIndex int) *big.Int {
	difficulty := new(big.Int)
	difficultyBytes := new([32]byte)
	difficultyBytes[difficultyIndex] = 1
	difficulty.SetBytes((*difficultyBytes)[:])
	return difficulty
}

func HashData(buf []byte) [32]byte {
	return sha256.Sum256(buf)
}
