package network

import (
    "edoc-blockchain/backend/internal/blockchain"
)

// MessageType определяет тип сообщения
type MessageType string

const (
    MsgNewTx    MessageType = "new_tx"    // новая транзакция
    MsgNewBlock MessageType = "new_block" // новый блок
)

// Message – общая структура для всех сообщений
type Message struct {
    Type MessageType `json:"type"`
    Data []byte      `json:"data"` // сериализованные данные (транзакция или блок)
}

// TxMessage – сообщение с транзакцией (удобная обёртка)
type TxMessage struct {
    Transaction *blockchain.Transaction `json:"transaction"`
}

// BlockMessage – сообщение с блоком
type BlockMessage struct {
    Block *blockchain.Block `json:"block"`
}