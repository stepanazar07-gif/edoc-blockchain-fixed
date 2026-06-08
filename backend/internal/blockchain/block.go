package blockchain

import (
	"bytes"
	"edoc-blockchain/backend/internal/blockchain/types"
	"edoc-blockchain/backend/internal/crypto"
	"encoding/gob"
	"fmt"
	"strings"
	"time"
)

// Header — служебная информация блока
type Header struct {
	Version       uint32     `json:"version"`
	PrevBlockHash types.Hash `json:"prev_block_hash"`
	Timestamp     int64      `json:"timestamp"`
	Height        uint32     `json:"height"`
	Nonce         uint64     `json:"nonce"`
}

// Block — один блок в цепочке
type Block struct {
	Header       *Header       `json:"header"`
	Transactions []Transaction `json:"transactions"`
	Hash         types.Hash    `json:"hash"`
}

// NewBlock создаёт новый блок (без майнинга)
func NewBlock(prevHash types.Hash, height uint32, txs []Transaction) *Block {
	header := &Header{
		Version:       1,
		PrevBlockHash: prevHash,
		Timestamp:     time.Now().UnixNano(),
		Height:        height,
		Nonce:         0,
	}
	block := &Block{
		Header:       header,
		Transactions: txs,
	}
	block.Hash = block.CalculateHash()
	return block
}

// CalculateHash вычисляет хэш блока
func (b *Block) CalculateHash() types.Hash {
	headerBytes, err := b.serializeHeader()
	if err != nil {
		return types.Hash{}
	}
	var txsData []byte
	for _, tx := range b.Transactions {
		txsData = append(txsData, tx.Hash.Bytes()...)
	}
	allData := append(headerBytes, txsData...)
	return crypto.HashData(allData)
}

// serializeHeader превращает заголовок в байты
func (b *Block) serializeHeader() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(b.Header); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Mine выполняет Proof-of-Work (майнинг)
func (b *Block) Mine(difficulty int) {
	target := strings.Repeat("0", difficulty)
	for {
		hash := b.CalculateHash()
		if strings.HasPrefix(hash.String(), target) {
			b.Hash = hash
			return
		}
		b.Header.Nonce++
	}
}

// Validate проверяет корректность блока
// Validate проверяет корректность блока
func (b *Block) Validate(prevBlock *Block) error {
    computedHash := b.CalculateHash()
    if !bytes.Equal(computedHash.Bytes(), b.Hash.Bytes()) {
        return fmt.Errorf("недопустимый хэш блока")
    }

    if prevBlock != nil {
        if b.Header.PrevBlockHash != prevBlock.Hash {
            return fmt.Errorf("несоответствие хэша предыдущего блока")
        }
        if b.Header.Height != prevBlock.Header.Height+1 {
            return fmt.Errorf("недопустимая высота")
        }
    }

    for _, tx := range b.Transactions {
        if !tx.Verify("mysecret") {   // ← здесь заменил "" на "mysecret"
            return fmt.Errorf("недействительная транзакция")
        }
    }

    return nil

}

// String для красивого вывода
func (b *Block) String() string {
	return fmt.Sprintf("Block{height:%d, hash:%s, txs:%d, nonce:%d}",
		b.Header.Height, b.Hash.String(), len(b.Transactions), b.Header.Nonce)
}
