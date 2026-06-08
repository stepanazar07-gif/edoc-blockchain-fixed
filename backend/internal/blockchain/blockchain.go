package blockchain

import (
    "fmt"
    "sync"
    "edoc-blockchain/backend/internal/blockchain/types"
)

// Blockchain представляет цепочку блоков, использующую хранилище
type Blockchain struct {
    store    Storage      // интерфейс хранилища (из этого же пакета)
    mu       sync.RWMutex // для синхронизации при майнинге/добавлении
    lastHash types.Hash   // кэш хэша последнего блока
}

// NewBlockchain создаёт цепочку с генезис-блоком и сохраняет его в хранилище
func NewBlockchain(store Storage) (*Blockchain, error) {
    bc := &Blockchain{
        store: store,
    }

    // Пытаемся получить последнюю высоту
    lastHeight, err := store.GetLastHeight()
    if err != nil {
        // Хранилище пустое – создаём и сохраняем генезис-блок
        genesis := bc.createGenesisBlock()
        if err := store.Put(genesis); err != nil {
            return nil, fmt.Errorf("failed to save genesis block: %w", err)
        }
        bc.lastHash = genesis.Hash
        return bc, nil
    }

    // Хранилище не пустое – загружаем последний блок и запоминаем его хэш
    lastBlock, err := store.Get(lastHeight)
    if err != nil {
        return nil, fmt.Errorf("failed to get last block: %w", err)
    }
    bc.lastHash = lastBlock.Hash
    return bc, nil
}

// createGenesisBlock создаёт первый блок (высота 0, PrevHash = ZeroHash)
func (bc *Blockchain) createGenesisBlock() *Block {
    return NewBlock(types.ZeroHash, 0, []Transaction{})
}

// AddBlock добавляет новый блок в цепочку (предполагается, что он уже намайнен)
func (bc *Blockchain) AddBlock(newBlock *Block) error {
    bc.mu.Lock()
    defer bc.mu.Unlock()

    // Проверяем, что новый блок ссылается на последний сохранённый
    if newBlock.Header.PrevBlockHash != bc.lastHash {
        return fmt.Errorf("previous hash mismatch")
    }

    // Получаем последнюю высоту из хранилища
    lastHeight, err := bc.store.GetLastHeight()
    if err != nil {
        return err
    }

    // Проверяем высоту
    if newBlock.Header.Height != lastHeight+1 {
        return fmt.Errorf("invalid height: expected %d, got %d",
            lastHeight+1, newBlock.Header.Height)
    }

    // Загружаем предыдущий блок для валидации
    prevBlock, err := bc.store.Get(lastHeight)
    if err != nil {
        return err
    }

    // Валидируем новый блок относительно предыдущего
    if err := newBlock.Validate(prevBlock); err != nil {
        return fmt.Errorf("block validation failed: %w", err)
    }

    // Сохраняем блок в хранилище
    if err := bc.store.Put(newBlock); err != nil {
        return err
    }

    // Обновляем кэш последнего хэша
    bc.lastHash = newBlock.Hash
    return nil
}

// LastBlock возвращает последний блок из хранилища
func (bc *Blockchain) LastBlock() (*Block, error) {
    lastHeight, err := bc.store.GetLastHeight()
    if err != nil {
        return nil, err
    }
    return bc.store.Get(lastHeight)
}

// GetBlock возвращает блок по высоте (номеру)
func (bc *Blockchain) GetBlock(height uint32) (*Block, error) {
    return bc.store.Get(height)
}

// GetLastHeight возвращает высоту последнего блока в цепочке
func (bc *Blockchain) GetLastHeight() (uint32, error) {
    return bc.store.GetLastHeight()
}

// ValidateChain проверяет целостность всей цепочки, начиная с генезис-блока
func (bc *Blockchain) ValidateChain() error {
    lastHeight, err := bc.store.GetLastHeight()
    if err != nil {
        return fmt.Errorf("failed to get last height: %w", err)
    }

    // Проверяем генезис-блок
    genesis, err := bc.store.Get(0)
    if err != nil {
        return fmt.Errorf("failed to get genesis block: %w", err)
    }
    if genesis.Header.Height != 0 {
        return fmt.Errorf("genesis block height must be 0, got %d", genesis.Header.Height)
    }
    if genesis.Header.PrevBlockHash != types.ZeroHash {
        return fmt.Errorf("genesis block prev hash must be zero, got %s", genesis.Header.PrevBlockHash.String())
    }

    // Проверяем каждый последующий блок
    for i := uint32(1); i <= lastHeight; i++ {
        block, err := bc.store.Get(i)
        if err != nil {
            return fmt.Errorf("failed to get block at height %d: %w", i, err)
        }
        prevBlock, err := bc.store.Get(i - 1)
        if err != nil {
            return fmt.Errorf("failed to get previous block at height %d: %w", i-1, err)
        }
        if err := block.Validate(prevBlock); err != nil {
            return fmt.Errorf("block %d is invalid: %w", i, err)
        }
    }
    return nil
}

// String возвращает строковое представление всей цепочки (для отладки)
func (bc *Blockchain) String() string {
    lastHeight, err := bc.store.GetLastHeight()
    if err != nil {
        return fmt.Sprintf("error: %v", err)
    }
    var result string
    for i := uint32(0); i <= lastHeight; i++ {
        block, err := bc.store.Get(i)
        if err != nil {
            result += fmt.Sprintf("error getting block %d: %v\n", i, err)
            continue
        }
        result += block.String() + "\n"
    }
    return result
}