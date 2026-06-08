package blockchain

import (
    "fmt"
    "sync"
)

type MemoryStorage struct {
    blocks map[uint32]*Block // ключ — высота блока (тип Block из этого же пакета)
    mu     sync.RWMutex
}

// NewMemoryStorage создаёт новое хранилище в памяти
func NewMemoryStorage() *MemoryStorage {
    return &MemoryStorage{
        blocks: make(map[uint32]*Block),
    }
}

// Put сохраняет блок в память
func (s *MemoryStorage) Put(block *Block) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.blocks[block.Header.Height] = block
    return nil
}

// Get возвращает блок по высоте
func (s *MemoryStorage) Get(height uint32) (*Block, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    block, ok := s.blocks[height]
    if !ok {
        return nil, fmt.Errorf("block at height %d not found", height)
    }
    return block, nil
}

// GetLastHeight возвращает максимальную высоту среди сохранённых блоков
func (s *MemoryStorage) GetLastHeight() (uint32, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    if len(s.blocks) == 0 {
        return 0, fmt.Errorf("no blocks in storage")
    }
    var maxHeight uint32
    for height := range s.blocks {
        if height > maxHeight {
            maxHeight = height
        }
    }
    return maxHeight, nil
}