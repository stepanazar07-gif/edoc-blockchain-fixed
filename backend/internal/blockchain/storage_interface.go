package blockchain

type Storage interface {
	Put(block *Block) error

	Get(height uint32) (*Block, error)

	GetLastHeight() (uint32, error)
}
