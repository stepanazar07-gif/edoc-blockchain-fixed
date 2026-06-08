package crypto

import (
	"crypto/sha256"
	"edoc-blockchain/backend/internal/blockchain/types"
)

func HashData(data []byte) types.Hash {
	hashBytes := sha256.Sum256(data)
	return types.Hash(hashBytes)
}
