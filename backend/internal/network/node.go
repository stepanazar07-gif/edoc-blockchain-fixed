package network

import (
	"edoc-blockchain/backend/internal/blockchain"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

// Node представляет участника P2P сети
type Node struct {
	bc       *blockchain.Blockchain // ссылка на наш блокчейн
	addr     string                 // адрес этого узла, например "localhost:8001"
	peers    map[string]net.Conn    // активные соединения с другими узлами (клч – адрес)
	mu       sync.Mutex             // защита карты peers
	listener net.Listener           // TCP-слушатель для входящих соединений
}

// NewNode создаёт новый узел, но не запускает его
func NewNode(bc *blockchain.Blockchain, addr string) *Node {
	return &Node{
		bc:    bc,
		addr:  addr,
		peers: make(map[string]net.Conn),
	}
}

// Start запускает TCP-сервер для приёма входящих соединений
func (n *Node) Start() error {
	listener, err := net.Listen("tcp", n.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	n.listener = listener
	fmt.Printf("P2P узел слушает на %s\n", n.addr)

	// Бесконечно принимаем соединения
	go n.acceptConnections()
	return nil
}

// acceptConnections – главный цикл приёма входящих соединений
func (n *Node) acceptConnections() {
	for {
		conn, err := n.listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		// Запускаем обработчик для каждого нового клиента (горутина)
		go n.handleConnection(conn)
	}
}

// handleConnection читает сообщения от подключившегося узла
func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()
	// Декодируем JSON-сообщения по одному
	decoder := json.NewDecoder(conn)
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			fmt.Println("Error decoding message:", err)
			return
		}
		n.processMessage(&msg)
	}
}

// processMessage обрабатывает полученное сообщение
func (n *Node) processMessage(msg *Message) {
	switch msg.Type {
	case MsgNewTx:
		var txMsg TxMessage
		if err := json.Unmarshal(msg.Data, &txMsg); err != nil {
			fmt.Println("Invalid tx message:", err)
			return
		}
		// Добавляем транзакцию в свой блокчейн (в mempool)
		// Для простоты просто выведем в консоль, а в реальном проекте нужен mempool
		fmt.Printf("Получена транзакция %s\n", txMsg.Transaction.Hash.String())
		// Здесь можно сохранить в mempool и при майнинге использовать

	case MsgNewBlock:
		var blockMsg BlockMessage
		if err := json.Unmarshal(msg.Data, &blockMsg); err != nil {
			fmt.Println("Invalid block message:", err)
			return
		}
		// Пытаемся добавить полученный блок в нашу цепочку
		// Сначала проверим валидность блока
		lastBlock, err := n.bc.LastBlock()
		if err != nil {
			fmt.Println("Can't get last block:", err)
			return
		}
		if err := blockMsg.Block.Validate(lastBlock); err != nil {
			fmt.Println("Invalid block received:", err)
			return
		}
		// Добавляем блок (без майнинга, он уже намайнен другим узлом)
		if err := n.bc.AddBlock(blockMsg.Block); err != nil {
			fmt.Println("Failed to add block:", err)
			return
		}
		fmt.Printf("Блок %d добавлен (получен от другого узла)\n", blockMsg.Block.Header.Height)

	default:
		fmt.Println("Unknown message type:", msg.Type)
	}
}

// Broadcast отправляет сообщение всем известным пирам
func (n *Node) Broadcast(msg *Message) error {
	// Сериализуем сообщение в JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	n.mu.Lock()
	// Перебираем копию пиров (чтобы не держать блокировку при отправке)
	peersCopy := make([]net.Conn, 0, len(n.peers))
	for _, conn := range n.peers {
		peersCopy = append(peersCopy, conn)
	}
	n.mu.Unlock()

	for _, conn := range peersCopy {
		// Добавляем символ новой строки в конце, чтобы decoder знал границу
		_, err := conn.Write(append(data, '\n'))
		if err != nil {
			fmt.Println("Failed to send to peer:", err)
			// Можно удалить пира из списка, но для простоты не будем
			continue
		}
	}
	return nil
}

// BroadcastTransaction рассылает транзакцию
func (n *Node) BroadcastTransaction(tx *blockchain.Transaction) error {
	txMsg := TxMessage{Transaction: tx}
	data, err := json.Marshal(txMsg)
	if err != nil {
		return err
	}
	return n.Broadcast(&Message{Type: MsgNewTx, Data: data})
}

// BroadcastBlock рассылает блок
func (n *Node) BroadcastBlock(block *blockchain.Block) error {
	blockMsg := BlockMessage{Block: block}
	data, err := json.Marshal(blockMsg)
	if err != nil {
		return err
	}
	return n.Broadcast(&Message{Type: MsgNewBlock, Data: data})
}
