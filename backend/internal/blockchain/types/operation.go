package types

import (
	"fmt"
	"time"
)

type OperationType byte

const (
	CreateOp OperationType = iota
	SignOp
	TransferOp
)

type DocumentOperation struct {
	Type      OperationType `json:"type"`
	Document  Document      `json:"document"`
	Signer    string        `json:"signer"`
	Timestamp int64         `json:"timestamp"`
}

func NewCreateOperation(doc Document, signer string) *DocumentOperation {
	return &DocumentOperation{
		Type:      CreateOp,
		Document:  doc,
		Signer:    signer,
		Timestamp: time.Now().UnixNano(),
	}
}

func NewSignOperation(doc Document, signer string) *DocumentOperation {
	return &DocumentOperation{
		Type:      SignOp, 
		Document:  doc,
		Signer:    signer,
		Timestamp: time.Now().UnixNano(),
	}
}

func NewTransferOperation(doc Document, signer string) *DocumentOperation {
	return &DocumentOperation{
		Type:      TransferOp,
		Document:  doc,
		Signer:    signer,
		Timestamp: time.Now().UnixNano(),
	}
}

func (op *DocumentOperation) String() string {
	opType := "неизвестный"
	switch op.Type {
	case CreateOp:
		opType = "творить"
	case SignOp:
		opType = "знак"
	case TransferOp:
		opType = "перемещение"
	}


	return fmt.Sprintf("Operation{type: %s, doc: %s, signer: %s, time: %d}",
		opType, op.Document.ID, op.Signer, op.Timestamp)
		
}


