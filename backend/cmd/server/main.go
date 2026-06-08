package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"edoc-blockchain/backend/internal/api"
	"edoc-blockchain/backend/internal/blockchain"
	"edoc-blockchain/backend/internal/network"
)

func main() {
	apiPort := getEnvInt("API_PORT", 8080)
	p2pAddr := getEnv("P2P_ADDR", "localhost:8001")
	connStr := getEnv("DATABASE_URL", "postgres://postgres:azar247911@localhost:5433/blockchain?sslmode=disable")

	// -------------------- ХРАНИЛИЩЕ И БЛОКЧЕЙН --------------------
	store, err := blockchain.NewPostgresStorage(connStr)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer store.Close()

	bc, err := blockchain.NewBlockchain(store)
	if err != nil {
		log.Fatalf("Blockchain initialization error: %v", err)
	}
	fmt.Println("Blockchain loaded from PostgreSQL")

	// -------------------- P2P УЗЕЛ (опционально) --------------------
	node := network.NewNode(bc, p2pAddr)
	if err := node.Start(); err != nil {
		log.Fatalf("P2P node start error: %v", err)
	}
	fmt.Printf("P2P node started on %s\n", p2pAddr)

	// -------------------- HTTP API (с поддержкой регистрации/логина) --------------------
	srv := api.NewServer(bc, store, apiPort)
	fmt.Printf("HTTP API configured on port %d\n", apiPort)

	// Запускаем API в отдельной горутине, чтобы не блокировать P2P
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("HTTP API error: %v", err)
		}
	}()

	select {}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
