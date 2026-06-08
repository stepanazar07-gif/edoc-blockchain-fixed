package types

import (
    "encoding/hex"
    "fmt"
)

type Hash [32]byte

func (h Hash) Bytes() []byte {
    return h[:]
}

func (h Hash) String() string {
    return hex.EncodeToString(h[:])
}

func (h Hash) IsZero() bool {
    for _, b := range h {
        if b != 0 {
            return false
        }
    }
    return true
}

// ZeroHash — пустой хэш (все 32 байта равны нулю)
var ZeroHash = Hash{}

func HashFromBytes(b []byte) (Hash, error) {
    if len(b) != 32 {
        return Hash{}, fmt.Errorf("invalid hash length: expected 32, got %d", len(b))
    }
    var h Hash
    copy(h[:], b)
    return h, nil
}

// HashFromString создаёт Hash из hex-строки
func HashFromString(s string) (Hash, error) {
    bytes, err := hex.DecodeString(s)
    if err != nil {
        return Hash{}, fmt.Errorf("invalid hex string: %w", err)
    }
    return HashFromBytes(bytes)
}