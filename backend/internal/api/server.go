package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"edoc-blockchain/backend/internal/auth"
	"edoc-blockchain/backend/internal/blockchain"
	"edoc-blockchain/backend/internal/blockchain/types"
)

const (
	maxDocumentSize = 50 << 20
	maxAvatarSize   = 5 << 20
)

type Server struct {
	bc        *blockchain.Blockchain
	store     *blockchain.PostgresStorage
	port      int
	uploadDir string
}

func NewServer(bc *blockchain.Blockchain, store *blockchain.PostgresStorage, port int) *Server {
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	return &Server{
		bc:        bc,
		store:     store,
		port:      port,
		uploadDir: uploadDir,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/status", s.statusHandler)
	http.HandleFunc("/register", s.registerHandler)
	http.HandleFunc("/login", s.loginHandler)

	http.HandleFunc("/document", auth.AuthMiddleware(s.documentHandler))
	http.HandleFunc("/document/", auth.AuthMiddleware(s.getDocumentHandler))
	http.HandleFunc("/download/", auth.AuthMiddleware(s.downloadHandler))
	http.HandleFunc("/me", auth.AuthMiddleware(s.meHandler))
	http.HandleFunc("/me/avatar", auth.AuthMiddleware(s.avatarHandler))
	http.HandleFunc("/users", auth.AuthMiddleware(s.usersHandler))
	http.HandleFunc("/my-documents", auth.AuthMiddleware(s.myDocumentsHandler))
	http.HandleFunc("/share-document", auth.AuthMiddleware(s.shareDocumentHandler))
	http.HandleFunc("/incoming-transfers", auth.AuthMiddleware(s.incomingTransfersHandler))
	http.HandleFunc("/accept-transfer", auth.AuthMiddleware(s.acceptTransferHandler))
	http.HandleFunc("/decline-transfer", auth.AuthMiddleware(s.declineTransferHandler))
	http.HandleFunc("/received-files", auth.AuthMiddleware(s.receivedFilesHandler))

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("HTTP API started on port %d\n", s.port)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Age      int    `json:"age"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Phone = normalizePhone(req.Phone)
	if req.Name == "" || req.Phone == "" || req.Password == "" {
		http.Error(w, "Name, phone and password are required", http.StatusBadRequest)
		return
	}
	if req.Age < 1 || req.Age > 120 {
		http.Error(w, "Age must be between 1 and 120", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		http.Error(w, "Password must contain at least 6 characters", http.StatusBadRequest)
		return
	}

	userID, err := s.store.CreateUser(req.Name, req.Age, req.Phone, req.Password)
	if err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	token, err := auth.GenerateToken(userID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"token": token,
		"id":    userID,
	})
}

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	phone := normalizePhone(req.Phone)
	if phone == "" || req.Password == "" {
		http.Error(w, "Phone and password are required", http.StatusBadRequest)
		return
	}

	user, hash, err := s.store.GetUserByPhone(phone)
	if err != nil || user == nil || !auth.CheckPasswordHash(req.Password, hash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token": token,
		"id":    user.ID,
	})
}

func (s *Server) documentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxDocumentSize)
	if err := r.ParseMultipartForm(maxDocumentSize); err != nil {
		http.Error(w, "File too large or invalid form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}
	if len(fileBytes) == 0 {
		http.Error(w, "File is empty", http.StatusBadRequest)
		return
	}

	hash := sha256.Sum256(fileBytes)
	contentHash := hex.EncodeToString(hash[:])
	fileName := filepath.Base(header.Filename)
	if formTitle := strings.TrimSpace(r.FormValue("title")); formTitle != "" {
		fileName = filepath.Base(formTitle)
	}
	mimeType := detectMimeType(fileName, fileBytes, header.Header.Get("Content-Type"))

	fileDir := filepath.Join(s.uploadDir, "files")
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}
	filePath := filepath.Join(fileDir, contentHash)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.WriteFile(filePath, fileBytes, 0600); err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
	}

	docID, err := s.store.CreateFile(userID, fileName, contentHash, filePath, int64(len(fileBytes)), mimeType)
	if err != nil {
		log.Printf("CreateFile error: %v", err)
		http.Error(w, "Failed to save file metadata", http.StatusInternalServerError)
		return
	}

	user, _ := s.store.GetUserByID(userID)
	owner := userID
	uploadedBy := ""
	if user != nil {
		owner = user.ID
		uploadedBy = user.Name
	}

	contentHashType, err := types.HashFromString(contentHash)
	if err != nil {
		http.Error(w, "Invalid content hash", http.StatusInternalServerError)
		return
	}
	doc := types.NewDocument(fileName, contentHashType, owner, int64(len(fileBytes)), mimeType)
	op := types.NewCreateOperation(*doc, owner)
	tx, err := blockchain.NewTransaction(op, owner)
	if err != nil {
		http.Error(w, "Failed to create transaction", http.StatusInternalServerError)
		return
	}
	if err := tx.Sign("coursework-demo-key"); err != nil {
		http.Error(w, "Failed to sign transaction", http.StatusInternalServerError)
		return
	}

	lastBlock, err := s.bc.LastBlock()
	if err != nil {
		http.Error(w, "Failed to get last block", http.StatusInternalServerError)
		return
	}
	newBlock := blockchain.NewBlock(lastBlock.Hash, lastBlock.Header.Height+1, []blockchain.Transaction{*tx})
	newBlock.Mine(2)
	if err := s.bc.AddBlock(newBlock); err != nil {
		http.Error(w, "Failed to add block: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status":           "ok",
		"transaction_hash": tx.Hash.String(),
		"block_hash":       newBlock.Hash.String(),
		"block_height":     newBlock.Header.Height,
		"document_id":      docID,
		"file_id":          docID,
		"file_name":        fileName,
		"file_hash":        contentHash,
		"file_size":        len(fileBytes),
		"mime_type":        mimeType,
		"uploaded_by":      uploadedBy,
	})
}

func (s *Server) getDocumentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hashHex := strings.TrimPrefix(r.URL.Path, "/document/")
	if hashHex == "" {
		http.Error(w, "Missing hash", http.StatusBadRequest)
		return
	}
	txHash, err := types.HashFromString(hashHex)
	if err != nil {
		http.Error(w, "Invalid hash format", http.StatusBadRequest)
		return
	}

	var foundTx *blockchain.Transaction
	var blockHeight uint32
	var blockHash types.Hash

	lastHeight, err := s.bc.GetLastHeight()
	if err != nil {
		http.Error(w, "Blockchain error", http.StatusInternalServerError)
		return
	}
	for h := uint32(0); h <= lastHeight; h++ {
		block, err := s.bc.GetBlock(h)
		if err != nil {
			continue
		}
		for _, tx := range block.Transactions {
			if tx.Hash == txHash {
				txCopy := tx
				foundTx = &txCopy
				blockHeight = block.Header.Height
				blockHash = block.Hash
				break
			}
		}
		if foundTx != nil {
			break
		}
	}
	if foundTx == nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	op, err := foundTx.DeserializeOperation()
	if err != nil {
		http.Error(w, "Invalid transaction data", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"transaction_hash": foundTx.Hash.String(),
		"block_height":     blockHeight,
		"block_hash":       blockHash.String(),
		"document":         op.Document,
		"operation_type":   op.Type,
		"signer":           op.Signer,
		"timestamp":        op.Timestamp,
	})
}

func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r)
	fileHash := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/download/"))
	if fileHash == "" {
		http.Error(w, "Missing file hash", http.StatusBadRequest)
		return
	}

	allowed, err := s.store.UserCanAccessFileHash(userID, fileHash)
	if err != nil || !allowed {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	file, err := s.store.GetFileByHash(fileHash)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	if _, err := os.Stat(file.FilePath); err != nil {
		http.Error(w, "Stored file is missing", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file.FileName))
	w.Header().Set("Content-Type", file.MimeType)
	http.ServeFile(w, r, file.FilePath)
}

func (s *Server) meHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r)
	user, err := s.store.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) avatarHandler(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)

	switch r.Method {
	case http.MethodGet:
		user, err := s.store.GetUserByID(userID)
		if err != nil || user.AvatarURL == "" {
			http.Error(w, "Avatar not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, user.AvatarURL)
	case http.MethodPost:
		s.uploadAvatar(w, r, userID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) uploadAvatar(w http.ResponseWriter, r *http.Request, userID string) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize)
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		http.Error(w, "Avatar is too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "Avatar file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	avatarBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read avatar", http.StatusInternalServerError)
		return
	}
	contentType := http.DetectContentType(avatarBytes)
	if contentType != "image/jpeg" && contentType != "image/png" {
		http.Error(w, "Only JPG and PNG avatars are allowed", http.StatusBadRequest)
		return
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(avatarBytes))
	if err != nil {
		http.Error(w, "Invalid image file", http.StatusBadRequest)
		return
	}
	if config.Width < 128 || config.Height < 128 || config.Width > 2048 || config.Height > 2048 {
		http.Error(w, "Avatar resolution must be from 128x128 to 2048x2048", http.StatusBadRequest)
		return
	}
	ratio := float64(config.Width) / float64(config.Height)
	if ratio < 0.5 || ratio > 2.0 {
		http.Error(w, "Avatar proportions are too extreme", http.StatusBadRequest)
		return
	}

	ext := ".jpg"
	if contentType == "image/png" {
		ext = ".png"
	}
	if header.Filename != "" {
		if detectedExt := strings.ToLower(filepath.Ext(header.Filename)); detectedExt == ".jpeg" {
			ext = ".jpg"
		}
	}

	avatarDir := filepath.Join(s.uploadDir, "avatars")
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		http.Error(w, "Failed to create avatar directory", http.StatusInternalServerError)
		return
	}
	avatarPath := filepath.Join(avatarDir, userID+ext)
	if err := os.WriteFile(avatarPath, avatarBytes, 0600); err != nil {
		http.Error(w, "Failed to save avatar", http.StatusInternalServerError)
		return
	}
	if err := s.store.UpdateUserAvatar(userID, avatarPath); err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"avatar_url": avatarPath})
}

func (s *Server) usersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r)
	searchID := strings.TrimSpace(r.URL.Query().Get("id"))
	if searchID != "" {
		user, err := s.store.GetPublicUserByID(searchID)
		if err != nil {
			writeJSON(w, http.StatusOK, []blockchain.PublicUser{})
			return
		}
		writeJSON(w, http.StatusOK, []blockchain.PublicUser{*user})
		return
	}

	users, err := s.store.GetAllPublicUsers(userID)
	if err != nil {
		http.Error(w, "Failed to load users", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (s *Server) myDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r)
	docs, err := s.store.GetUserFiles(userID)
	if err != nil {
		http.Error(w, "Failed to load files: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, docs)
}

func (s *Server) shareDocumentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r)
	var req struct {
		FileID     string `json:"file_id"`
		DocumentID string `json:"document_id"`
		ReceiverID string `json:"receiver_id"`
		ToUserID   string `json:"to_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.FileID == "" {
		req.FileID = req.DocumentID
	}
	if req.ReceiverID == "" {
		req.ReceiverID = req.ToUserID
	}
	if req.FileID == "" || req.ReceiverID == "" {
		http.Error(w, "file_id and receiver_id are required", http.StatusBadRequest)
		return
	}
	if req.ReceiverID == userID {
		http.Error(w, "Cannot send a file to yourself", http.StatusBadRequest)
		return
	}

	file, err := s.store.GetFileByID(req.FileID)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	if file.OwnerID != userID {
		http.Error(w, "Only owner can send this file", http.StatusForbidden)
		return
	}
	if _, err := s.store.GetPublicUserByID(req.ReceiverID); err != nil {
		http.Error(w, "Receiver not found", http.StatusNotFound)
		return
	}

	transferID, err := s.store.CreateTransfer(req.FileID, userID, req.ReceiverID)
	if err != nil {
		http.Error(w, "Failed to create transfer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"transfer_id": transferID,
		"file_hash":   file.FileHash,
	})
}

func (s *Server) incomingTransfersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	transfers, err := s.store.GetIncomingTransfers(auth.GetUserID(r))
	if err != nil {
		http.Error(w, "Failed to load transfers: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, transfers)
}

func (s *Server) acceptTransferHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r)
	var req struct {
		TransferID string `json:"transfer_id"`
		FileHash   string `json:"file_hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	req.TransferID = strings.TrimSpace(req.TransferID)
	req.FileHash = strings.ToLower(strings.TrimSpace(req.FileHash))
	if req.TransferID == "" || req.FileHash == "" {
		http.Error(w, "transfer_id and file_hash are required", http.StatusBadRequest)
		return
	}

	transfer, file, ok := s.loadPendingReceiverTransfer(w, req.TransferID, userID)
	if !ok {
		return
	}
	if !strings.EqualFold(file.FileHash, req.FileHash) {
		http.Error(w, "Hash mismatch", http.StatusForbidden)
		return
	}
	if err := s.store.UpdateTransferStatus(transfer.ID, "accepted"); err != nil {
		http.Error(w, "Failed to update transfer", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"transfer_id": transfer.ID,
		"file_id":     file.ID,
		"file_name":   file.FileName,
		"file_hash":   file.FileHash,
		"file_size":   file.FileSize,
		"mime_type":   file.MimeType,
		"sender_id":   transfer.SenderID,
		"status":      "accepted",
	})
}

func (s *Server) declineTransferHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := auth.GetUserID(r)
	var req struct {
		TransferID string `json:"transfer_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	req.TransferID = strings.TrimSpace(req.TransferID)
	if req.TransferID == "" {
		http.Error(w, "transfer_id is required", http.StatusBadRequest)
		return
	}

	transfer, _, ok := s.loadPendingReceiverTransfer(w, req.TransferID, userID)
	if !ok {
		return
	}
	if err := s.store.UpdateTransferStatus(transfer.ID, "declined"); err != nil {
		http.Error(w, "Failed to update transfer", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"transfer_id": transfer.ID,
		"status":      "declined",
	})
}

func (s *Server) receivedFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	files, err := s.store.GetReceivedFiles(auth.GetUserID(r))
	if err != nil {
		http.Error(w, "Failed to load received files", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, files)
}

func (s *Server) loadPendingReceiverTransfer(w http.ResponseWriter, transferID string, userID string) (*blockchain.Transfer, *blockchain.FileRecord, bool) {
	transfer, err := s.store.GetTransferByID(transferID)
	if err != nil {
		http.Error(w, "Transfer not found", http.StatusNotFound)
		return nil, nil, false
	}
	if transfer.ReceiverID != userID {
		http.Error(w, "This transfer is not for you", http.StatusForbidden)
		return nil, nil, false
	}
	if transfer.Status != "pending" {
		http.Error(w, "Transfer is not pending", http.StatusGone)
		return nil, nil, false
	}
	file, err := s.store.GetFileByID(transfer.FileID)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return nil, nil, false
	}
	return transfer, file, true
}

func normalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	return phone
}

func detectMimeType(fileName string, fileBytes []byte, headerType string) string {
	if headerType != "" && headerType != "application/octet-stream" {
		return headerType
	}
	if ext := filepath.Ext(fileName); ext != "" {
		if byExt := mime.TypeByExtension(ext); byExt != "" {
			return byExt
		}
	}
	return http.DetectContentType(fileBytes)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
