package blockchain

import (
	"bytes"
	"edoc-blockchain/backend/internal/blockchain/types"
	"edoc-blockchain/backend/internal/crypto"
	"encoding/gob"
	"fmt"
	"time"
)

// Transaction — это запись о действии в блокчейне
type Transaction struct {
	Data      []byte     `json:"data"`
	From      string     `json:"from"`
	Signature []byte     `json:"signature"`
	Hash      types.Hash `json:"hash"`
	FirstSeen int64      `json:"first_seen"`
}

// serializeOperation превращает операцию в байты
func (tx *Transaction) serializeOperation(op *types.DocumentOperation) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(op); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// deserializeOperation превращает байты обратно в операцию
// DeserializeOperation превращает Data обратно в DocumentOperation
func (tx *Transaction) DeserializeOperation() (*types.DocumentOperation, error) {
    buf := bytes.NewReader(tx.Data)
    decoder := gob.NewDecoder(buf)
    var op types.DocumentOperation
    if err := decoder.Decode(&op); err != nil {
        return nil, err
    }
    return &op, nil
}

// NewTransaction создаёт новую транзакцию из операции
func NewTransaction(op *types.DocumentOperation, from string) (*Transaction, error) {
	tx := &Transaction{
		From:      from,
		Signature: []byte{},
		FirstSeen: time.Now().UnixNano(),
	}

	// Сериализуем операцию в байты
	data, err := tx.serializeOperation(op)
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить операцию сериализации: %w", err)
	}
	tx.Data = data

	// Вычисляем хэш
	tx.Hash = tx.CalculateHash()

	return tx, nil
}
 
// CalculateHash вычисляет SHA-256 хэш транзакции
func (tx *Transaction) CalculateHash() types.Hash {
	var data []byte
	data = append(data, tx.Data...)
	data = append(data, []byte(tx.From)...)
	timeBytes := []byte(fmt.Sprintf("%d", tx.FirstSeen))
	data = append(data, timeBytes...)
	return crypto.HashData(data)
}

// Sign подписывает транзакцию (упрощённая версия)
func (tx *Transaction) Sign(privateKey string) error {
	dataToSign := append(tx.Data, []byte(tx.From)...)
	dataToSign = append(dataToSign, []byte(privateKey)...)
	hash := crypto.HashData(dataToSign)
	tx.Signature = hash.Bytes()
	return nil
}

// Verify проверяет подпись транзакции
func (tx *Transaction) Verify(publicKey string) bool {
	dataToSign := append(tx.Data, []byte(tx.From)...)
	dataToSign = append(dataToSign, []byte(publicKey)...)
	expectedHash := crypto.HashData(dataToSign)
	return bytes.Equal(tx.Signature, expectedHash.Bytes())
}

// String для красивого вывода
func (tx *Transaction) String() string {
	return fmt.Sprintf("Tx{hash: %s, from: %s, firstSeen: %d, dataLen: %d}",
		tx.Hash.String(), tx.From, tx.FirstSeen, len(tx.Data))
}
