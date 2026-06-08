package blockchain

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"edoc-blockchain/backend/internal/auth"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	Phone     string    `json:"phone"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

type PublicUser struct {
	ID string `json:"id"`
}

type FileRecord struct {
	ID             string `json:"id"`
	OwnerID        string `json:"owner_id"`
	FileName       string `json:"file_name"`
	FileHash       string `json:"file_hash"`
	FilePath       string `json:"file_path,omitempty"`
	MimeType       string `json:"mime_type"`
	FileSize       int64  `json:"file_size"`
	UploadDate     string `json:"upload_date"`
	UploadTime     string `json:"upload_time"`
	UploadedByName string `json:"uploaded_by"`
}

type Transfer struct {
	ID         string `json:"id"`
	FileID     string `json:"file_id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	Status     string `json:"status"`
}

type TransferView struct {
	TransferID   string `json:"transfer_id"`
	FileID       string `json:"file_id"`
	FileName     string `json:"file_name"`
	FileHash     string `json:"file_hash,omitempty"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
	SenderID     string `json:"sender_id"`
	ReceiverID   string `json:"receiver_id"`
	Status       string `json:"status"`
	TransferDate string `json:"transfer_date"`
	AcceptedAt   string `json:"accepted_at,omitempty"`
}

func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}
	return &PostgresStorage{db: db}, nil
}

func createTables(db *sql.DB) error {
	query := `
	CREATE EXTENSION IF NOT EXISTS pgcrypto;

	CREATE TABLE IF NOT EXISTS blocks (
		height INTEGER PRIMARY KEY,
		block_data JSONB NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS transactions (
		tx_hash TEXT PRIMARY KEY,
		block_height INTEGER NOT NULL REFERENCES blocks(height) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_tx_hash ON transactions(tx_hash);

	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name TEXT NOT NULL,
		age INTEGER NOT NULL CHECK (age BETWEEN 1 AND 120),
		phone VARCHAR(32) UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		avatar_url TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	ALTER TABLE users ADD COLUMN IF NOT EXISTS name TEXT;
	ALTER TABLE users ADD COLUMN IF NOT EXISTS age INTEGER;
	ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(32);
	ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT;

	DO $$
	DECLARE
		has_username BOOLEAN;
	BEGIN
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'username'
		) INTO has_username;

		IF has_username THEN
			EXECUTE 'ALTER TABLE users ALTER COLUMN username DROP NOT NULL';
			EXECUTE 'UPDATE users SET name = COALESCE(NULLIF(name, ''''), NULLIF(username, ''''), ''User'') WHERE name IS NULL OR name = ''''';
			EXECUTE 'UPDATE users SET phone = LEFT(COALESCE(NULLIF(phone, ''''), NULLIF(username, ''''), id::text), 32) WHERE phone IS NULL OR phone = ''''';
		ELSE
			UPDATE users SET name = COALESCE(NULLIF(name, ''), 'User') WHERE name IS NULL OR name = '';
			UPDATE users SET phone = LEFT(COALESCE(NULLIF(phone, ''), id::text), 32) WHERE phone IS NULL OR phone = '';
		END IF;

		UPDATE users SET age = 18 WHERE age IS NULL OR age < 1 OR age > 120;
	END $$;

	ALTER TABLE users ALTER COLUMN name SET NOT NULL;
	ALTER TABLE users ALTER COLUMN age SET NOT NULL;
	ALTER TABLE users ALTER COLUMN phone SET NOT NULL;
	CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_unique ON users(phone);

	CREATE TABLE IF NOT EXISTS files (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		file_name TEXT NOT NULL,
		file_hash TEXT NOT NULL,
		file_path TEXT NOT NULL,
		upload_date DATE NOT NULL DEFAULT CURRENT_DATE,
		upload_time TIME NOT NULL DEFAULT CURRENT_TIME,
		file_size BIGINT NOT NULL CHECK (file_size >= 0),
		mime_type TEXT NOT NULL DEFAULT 'application/octet-stream'
	);
	CREATE INDEX IF NOT EXISTS idx_files_owner_id ON files(owner_id);
	CREATE INDEX IF NOT EXISTS idx_files_file_hash ON files(file_hash);

	CREATE TABLE IF NOT EXISTS file_transfers (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
		sender_id UUID NOT NULL REFERENCES users(id),
		receiver_id UUID NOT NULL REFERENCES users(id),
		status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined')),
		transfer_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		accepted_at TIMESTAMP WITH TIME ZONE,
		CHECK (sender_id <> receiver_id)
	);
	CREATE INDEX IF NOT EXISTS idx_file_transfers_receiver_id ON file_transfers(receiver_id);
	CREATE INDEX IF NOT EXISTS idx_file_transfers_sender_id ON file_transfers(sender_id);
	CREATE INDEX IF NOT EXISTS idx_file_transfers_status ON file_transfers(status);

	DO $$
	BEGIN
		IF to_regclass('public.documents') IS NOT NULL THEN
			INSERT INTO files (id, owner_id, file_name, file_hash, file_path, upload_date, upload_time, file_size, mime_type)
			SELECT id, owner_id, file_name, file_hash, './uploads/' || file_hash,
			       uploaded_at::date, uploaded_at::time,
			       file_size, COALESCE(mime_type, 'application/octet-stream')
			FROM documents
			ON CONFLICT (id) DO NOTHING;
		END IF;

		IF to_regclass('public.transfers') IS NOT NULL THEN
			INSERT INTO file_transfers (id, file_id, sender_id, receiver_id, status, transfer_date, accepted_at)
			SELECT id, document_id, from_user_id, to_user_id,
			       CASE
			           WHEN status IN ('claimed', 'accepted') THEN 'accepted'
			           WHEN status = 'declined' THEN 'declined'
			           ELSE 'pending'
			       END,
			       created_at,
			       claimed_at
			FROM transfers
			ON CONFLICT (id) DO NOTHING;
		END IF;
	END $$;
	`
	_, err := db.Exec(query)
	return err
}

func (s *PostgresStorage) Put(block *Block) error {
	blockJSON, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(
		`INSERT INTO blocks (height, block_data) VALUES ($1, $2)
		 ON CONFLICT (height) DO UPDATE SET block_data = EXCLUDED.block_data`,
		block.Header.Height, blockJSON,
	)
	if err != nil {
		return fmt.Errorf("insert block: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM transactions WHERE block_height = $1`, block.Header.Height)
	if err != nil {
		return err
	}
	for _, t := range block.Transactions {
		_, err = tx.Exec(`INSERT INTO transactions (tx_hash, block_height) VALUES ($1, $2)`,
			t.Hash.String(), block.Header.Height)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PostgresStorage) Get(height uint32) (*Block, error) {
	var data []byte
	err := s.db.QueryRow(`SELECT block_data FROM blocks WHERE height = $1`, height).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("block at height %d not found", height)
	}
	if err != nil {
		return nil, err
	}
	var block Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &block, nil
}

func (s *PostgresStorage) GetLastHeight() (uint32, error) {
	var max uint32
	err := s.db.QueryRow(`SELECT COALESCE(MAX(height), 0) FROM blocks`).Scan(&max)
	if err != nil {
		return 0, err
	}
	if max == 0 {
		var count int
		err = s.db.QueryRow(`SELECT COUNT(*) FROM blocks WHERE height = 0`).Scan(&count)
		if err != nil {
			return 0, err
		}
		if count == 0 {
			return 0, fmt.Errorf("no blocks in storage")
		}
	}
	return max, nil
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

func (s *PostgresStorage) CreateUser(name string, age int, phone string, password string) (string, error) {
	hashed, err := auth.HashPassword(password)
	if err != nil {
		return "", err
	}
	var userID string
	err = s.db.QueryRow(
		`INSERT INTO users (name, age, phone, password_hash) VALUES ($1, $2, $3, $4) RETURNING id`,
		name, age, phone, hashed,
	).Scan(&userID)
	return userID, err
}

func (s *PostgresStorage) GetUserByPhone(phone string) (*User, string, error) {
	var user User
	var passwordHash string
	err := s.db.QueryRow(`
		SELECT id, name, age, phone, password_hash, COALESCE(avatar_url, ''), created_at
		FROM users WHERE phone = $1`,
		phone,
	).Scan(&user.ID, &user.Name, &user.Age, &user.Phone, &passwordHash, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", nil
		}
		return nil, "", err
	}
	return &user, passwordHash, nil
}

func (s *PostgresStorage) GetUserByID(id string) (*User, error) {
	var user User
	err := s.db.QueryRow(`
		SELECT id, name, age, phone, COALESCE(avatar_url, ''), created_at
		FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Name, &user.Age, &user.Phone, &user.AvatarURL, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *PostgresStorage) GetPublicUserByID(id string) (*PublicUser, error) {
	var user PublicUser
	err := s.db.QueryRow(`SELECT id FROM users WHERE id = $1`, id).Scan(&user.ID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *PostgresStorage) GetAllPublicUsers(excludeUserID string) ([]PublicUser, error) {
	rows, err := s.db.Query(`SELECT id FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]PublicUser, 0)
	for rows.Next() {
		var user PublicUser
		if err := rows.Scan(&user.ID); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *PostgresStorage) UpdateUserAvatar(userID string, avatarURL string) error {
	_, err := s.db.Exec(`UPDATE users SET avatar_url = $1 WHERE id = $2`, avatarURL, userID)
	return err
}

func (s *PostgresStorage) CreateFile(ownerID string, fileName string, fileHash string, filePath string, fileSize int64, mimeType string) (string, error) {
	var fileID string
	err := s.db.QueryRow(`
		INSERT INTO files (owner_id, file_name, file_hash, file_path, file_size, mime_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		ownerID, fileName, fileHash, filePath, fileSize, mimeType,
	).Scan(&fileID)
	return fileID, err
}

func (s *PostgresStorage) GetUserFiles(userID string) ([]FileRecord, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.owner_id, f.file_name, f.file_hash, f.file_path, f.file_size, f.mime_type,
		       f.upload_date::text, to_char(f.upload_time, 'HH24:MI:SS'), u.name
		FROM files f
		JOIN users u ON u.id = f.owner_id
		WHERE f.owner_id = $1
		ORDER BY f.upload_date DESC, f.upload_time DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFileRecords(rows)
}

func (s *PostgresStorage) GetFileByID(fileID string) (*FileRecord, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.owner_id, f.file_name, f.file_hash, f.file_path, f.file_size, f.mime_type,
		       f.upload_date::text, to_char(f.upload_time, 'HH24:MI:SS'), u.name
		FROM files f
		JOIN users u ON u.id = f.owner_id
		WHERE f.id = $1`,
		fileID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files, err := scanFileRecords(rows)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, sql.ErrNoRows
	}
	return &files[0], nil
}

func (s *PostgresStorage) GetFileByHash(fileHash string) (*FileRecord, error) {
	rows, err := s.db.Query(`
		SELECT f.id, f.owner_id, f.file_name, f.file_hash, f.file_path, f.file_size, f.mime_type,
		       f.upload_date::text, to_char(f.upload_time, 'HH24:MI:SS'), u.name
		FROM files f
		JOIN users u ON u.id = f.owner_id
		WHERE f.file_hash = $1
		ORDER BY f.upload_date DESC, f.upload_time DESC
		LIMIT 1`,
		fileHash,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files, err := scanFileRecords(rows)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, sql.ErrNoRows
	}
	return &files[0], nil
}

func (s *PostgresStorage) UserCanAccessFileHash(userID string, fileHash string) (bool, error) {
	var allowed bool
	err := s.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM files
			WHERE file_hash = $1 AND owner_id = $2
			UNION
			SELECT 1
			FROM file_transfers t
			JOIN files f ON f.id = t.file_id
			WHERE f.file_hash = $1 AND t.receiver_id = $2 AND t.status = 'accepted'
		)`,
		fileHash, userID,
	).Scan(&allowed)
	return allowed, err
}

func (s *PostgresStorage) CreateTransfer(fileID string, senderID string, receiverID string) (string, error) {
	var transferID string
	err := s.db.QueryRow(`
		INSERT INTO file_transfers (file_id, sender_id, receiver_id)
		VALUES ($1, $2, $3)
		RETURNING id`,
		fileID, senderID, receiverID,
	).Scan(&transferID)
	return transferID, err
}

func (s *PostgresStorage) GetTransferByID(transferID string) (*Transfer, error) {
	var transfer Transfer
	err := s.db.QueryRow(`
		SELECT id, file_id, sender_id, receiver_id, status
		FROM file_transfers WHERE id = $1`,
		transferID,
	).Scan(&transfer.ID, &transfer.FileID, &transfer.SenderID, &transfer.ReceiverID, &transfer.Status)
	if err != nil {
		return nil, err
	}
	return &transfer, nil
}

func (s *PostgresStorage) UpdateTransferStatus(transferID string, status string) error {
	if status == "accepted" {
		_, err := s.db.Exec(
			`UPDATE file_transfers SET status = $1, accepted_at = NOW() WHERE id = $2`,
			status, transferID,
		)
		return err
	}
	_, err := s.db.Exec(`UPDATE file_transfers SET status = $1 WHERE id = $2`, status, transferID)
	return err
}

func (s *PostgresStorage) GetIncomingTransfers(userID string) ([]TransferView, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.file_id, f.file_name, f.file_size, f.mime_type, t.sender_id, t.receiver_id,
		       t.status, to_char(t.transfer_date, 'YYYY-MM-DD HH24:MI:SS')
		FROM file_transfers t
		JOIN files f ON f.id = t.file_id
		WHERE t.receiver_id = $1 AND t.status = 'pending'
		ORDER BY t.transfer_date DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transfers := make([]TransferView, 0)
	for rows.Next() {
		var transfer TransferView
		if err := rows.Scan(
			&transfer.TransferID,
			&transfer.FileID,
			&transfer.FileName,
			&transfer.FileSize,
			&transfer.MimeType,
			&transfer.SenderID,
			&transfer.ReceiverID,
			&transfer.Status,
			&transfer.TransferDate,
		); err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}
	return transfers, rows.Err()
}

func (s *PostgresStorage) GetReceivedFiles(userID string) ([]TransferView, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.file_id, f.file_name, f.file_hash, f.file_size, f.mime_type,
		       t.sender_id, t.receiver_id, t.status,
		       to_char(t.transfer_date, 'YYYY-MM-DD HH24:MI:SS'),
		       COALESCE(to_char(t.accepted_at, 'YYYY-MM-DD HH24:MI:SS'), '')
		FROM file_transfers t
		JOIN files f ON f.id = t.file_id
		WHERE t.receiver_id = $1 AND t.status = 'accepted'
		ORDER BY t.accepted_at DESC NULLS LAST, t.transfer_date DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transfers := make([]TransferView, 0)
	for rows.Next() {
		var transfer TransferView
		if err := rows.Scan(
			&transfer.TransferID,
			&transfer.FileID,
			&transfer.FileName,
			&transfer.FileHash,
			&transfer.FileSize,
			&transfer.MimeType,
			&transfer.SenderID,
			&transfer.ReceiverID,
			&transfer.Status,
			&transfer.TransferDate,
			&transfer.AcceptedAt,
		); err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}
	return transfers, rows.Err()
}

func scanFileRecords(rows *sql.Rows) ([]FileRecord, error) {
	files := make([]FileRecord, 0)
	for rows.Next() {
		var file FileRecord
		if err := rows.Scan(
			&file.ID,
			&file.OwnerID,
			&file.FileName,
			&file.FileHash,
			&file.FilePath,
			&file.FileSize,
			&file.MimeType,
			&file.UploadDate,
			&file.UploadTime,
			&file.UploadedByName,
		); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, rows.Err()
}
