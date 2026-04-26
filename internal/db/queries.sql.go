package db

import (
	"context"
	"database/sql"
)

type Querier interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type Queries struct {
	db Querier
}

func NewQueries(db *sql.DB) *Queries {
	return &Queries{db: db}
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{db: tx}
}

// Chat queries

func (q *Queries) CreateChat(ctx context.Context, arg CreateChatParams) (*Chat, error) {
	row := q.db.QueryRowContext(ctx, createChat,
		arg.ID, arg.ParentChatID, arg.Title, arg.Metadata,
	)
	var c Chat
	err := row.Scan(
		&c.ID, &c.ParentChatID, &c.Title, &c.MessageCount,
		&c.PromptTokens, &c.CompletionTokens, &c.Cost,
		&c.SummaryMessageID, &c.Metadata, &c.CreatedAt, &c.UpdatedAt,
	)
	return &c, err
}

type CreateChatParams struct {
	ID           string
	ParentChatID *string
	Title       string
	Metadata   string
}

func (q *Queries) GetChatByID(ctx context.Context, id string) (*Chat, error) {
	row := q.db.QueryRowContext(ctx, getChatByID, id)
	var c Chat
	err := row.Scan(
		&c.ID, &c.ParentChatID, &c.Title, &c.MessageCount,
		&c.PromptTokens, &c.CompletionTokens, &c.Cost,
		&c.SummaryMessageID, &c.Metadata, &c.CreatedAt, &c.UpdatedAt,
	)
	return &c, err
}

func (q *Queries) GetAllChats(ctx context.Context) ([]Chat, error) {
	rows, err := q.db.QueryContext(ctx, getAllChats)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chats []Chat
	for rows.Next() {
		var c Chat
		err := rows.Scan(
			&c.ID, &c.ParentChatID, &c.Title, &c.MessageCount,
			&c.PromptTokens, &c.CompletionTokens, &c.Cost,
			&c.SummaryMessageID, &c.Metadata, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		chats = append(chats, c)
	}
	return chats, rows.Err()
}

func (q *Queries) GetChatsByParentID(ctx context.Context, parentChatID string) ([]Chat, error) {
	rows, err := q.db.QueryContext(ctx, getChatsByParentID, parentChatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chats []Chat
	for rows.Next() {
		var c Chat
		err := rows.Scan(
			&c.ID, &c.ParentChatID, &c.Title, &c.MessageCount,
			&c.PromptTokens, &c.CompletionTokens, &c.Cost,
			&c.SummaryMessageID, &c.Metadata, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		chats = append(chats, c)
	}
	return chats, rows.Err()
}

func (q *Queries) UpdateChatTitle(ctx context.Context, arg UpdateChatTitleParams) error {
	_, err := q.db.ExecContext(ctx, updateChatTitle, arg.Title, arg.ID)
	return err
}

type UpdateChatTitleParams struct {
	ID    string
	Title string
}

func (q *Queries) UpdateChatMetadata(ctx context.Context, arg UpdateChatMetadataParams) error {
	_, err := q.db.ExecContext(ctx, updateChatMetadata, arg.Metadata, arg.ID)
	return err
}

type UpdateChatMetadataParams struct {
	ID        string
	Metadata string
}

func (q *Queries) UpdateChatStats(ctx context.Context, arg UpdateChatStatsParams) error {
	_, err := q.db.ExecContext(ctx, updateChatStats,
		arg.ID, arg.PromptTokens, arg.CompletionTokens, arg.Cost,
	)
	return err
}

type UpdateChatStatsParams struct {
	ID               string
	PromptTokens    int
	CompletionTokens int
	Cost           float64
}

func (q *Queries) SetChatSummary(ctx context.Context, arg SetChatSummaryParams) error {
	_, err := q.db.ExecContext(ctx, setChatSummary, arg.ID, arg.SummaryMessageID)
	return err
}

type SetChatSummaryParams struct {
	ID               string
	SummaryMessageID *string
}

func (q *Queries) DeleteChat(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteChat, id)
	return err
}

// Message queries

func (q *Queries) CreateMessage(ctx context.Context, arg CreateMessageParams) (*Message, error) {
	row := q.db.QueryRowContext(ctx, createMessage,
		arg.ID, arg.ChatID, arg.Role, arg.Content,
		arg.Model, arg.Provider, arg.IsSummary,
	)
	var m Message
	err := row.Scan(
		&m.ID, &m.ChatID, &m.Role, &m.Content,
		&m.Model, &m.Provider, &m.IsSummary,
		&m.CreatedAt, &m.UpdatedAt, &m.FinishedAt,
	)
	return &m, err
}

type CreateMessageParams struct {
	ID        string
	ChatID    string
	Role      Role
	Content   string
	Model    *string
	Provider *string
	IsSummary bool
}

func (q *Queries) GetMessageByID(ctx context.Context, id string) (*Message, error) {
	row := q.db.QueryRowContext(ctx, getMessageByID, id)
	var m Message
	err := row.Scan(
		&m.ID, &m.ChatID, &m.Role, &m.Content,
		&m.Model, &m.Provider, &m.IsSummary,
		&m.CreatedAt, &m.UpdatedAt, &m.FinishedAt,
	)
	return &m, err
}

func (q *Queries) GetMessagesByChatID(ctx context.Context, chatID string) ([]Message, error) {
	rows, err := q.db.QueryContext(ctx, getMessagesByChatID, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []Message
	for rows.Next() {
		var m Message
		err := rows.Scan(
			&m.ID, &m.ChatID, &m.Role, &m.Content,
			&m.Model, &m.Provider, &m.IsSummary,
			&m.CreatedAt, &m.UpdatedAt, &m.FinishedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (q *Queries) GetMessagesByChatIDPaginated(ctx context.Context, arg GetMessagesByChatIDPaginatedParams) ([]Message, error) {
	rows, err := q.db.QueryContext(ctx, getMessagesByChatIDPaginated, arg.ChatID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []Message
	for rows.Next() {
		var m Message
		err := rows.Scan(
			&m.ID, &m.ChatID, &m.Role, &m.Content,
			&m.Model, &m.Provider, &m.IsSummary,
			&m.CreatedAt, &m.UpdatedAt, &m.FinishedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

type GetMessagesByChatIDPaginatedParams struct {
	ChatID string
	Limit int
	Offset int
}

func (q *Queries) GetChatLatestMessage(ctx context.Context, chatID string) (*Message, error) {
	row := q.db.QueryRowContext(ctx, getChatLatestMessage, chatID)
	var m Message
	err := row.Scan(
		&m.ID, &m.ChatID, &m.Role, &m.Content,
		&m.Model, &m.Provider, &m.IsSummary,
		&m.CreatedAt, &m.UpdatedAt, &m.FinishedAt,
	)
	return &m, err
}

func (q *Queries) UpdateMessageContent(ctx context.Context, arg UpdateMessageContentParams) error {
	_, err := q.db.ExecContext(ctx, updateMessageContent, arg.ID, arg.Content)
	return err
}

type UpdateMessageContentParams struct {
	ID      string
	Content string
}

func (q *Queries) UpdateMessageFinishedAt(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, updateMessageFinishedAt, id)
	return err
}

func (q *Queries) DeleteMessage(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteMessage, id)
	return err
}

func (q *Queries) DeleteMessagesByChatID(ctx context.Context, chatID string) error {
	_, err := q.db.ExecContext(ctx, deleteMessagesByChatID, chatID)
	return err
}

func (q *Queries) CountMessagesByChatID(ctx context.Context, chatID string) (int, error) {
	row := q.db.QueryRowContext(ctx, countMessagesByChatID, chatID)
	var count int
	err := row.Scan(&count)
	return count, err
}

// File queries

func (q *Queries) CreateFile(ctx context.Context, arg CreateFileParams) (*File, error) {
	row := q.db.QueryRowContext(ctx, createFile,
		arg.ID, arg.ChatID, arg.Path,
	)
	var f File
	err := row.Scan(
		&f.ID, &f.ChatID, &f.Path, &f.Version,
		&f.CreatedAt, &f.UpdatedAt,
	)
	return &f, err
}

type CreateFileParams struct {
	ID    string
	ChatID string
	Path  string
}

func (q *Queries) GetFileByID(ctx context.Context, id string) (*File, error) {
	row := q.db.QueryRowContext(ctx, getFileByID, id)
	var f File
	err := row.Scan(
		&f.ID, &f.ChatID, &f.Path, &f.Version,
		&f.CreatedAt, &f.UpdatedAt,
	)
	return &f, err
}

func (q *Queries) GetFileByChatAndPath(ctx context.Context, arg GetFileByChatAndPathParams) (*File, error) {
	row := q.db.QueryRowContext(ctx, getFileByChatAndPath, arg.ChatID, arg.Path)
	var f File
	err := row.Scan(
		&f.ID, &f.ChatID, &f.Path, &f.Version,
		&f.CreatedAt, &f.UpdatedAt,
	)
	return &f, err
}

type GetFileByChatAndPathParams struct {
	ChatID string
	Path  string
}

func (q *Queries) GetFilesByChatID(ctx context.Context, chatID string) ([]File, error) {
	rows, err := q.db.QueryContext(ctx, getFilesByChatID, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files []File
	for rows.Next() {
		var f File
		err := rows.Scan(
			&f.ID, &f.ChatID, &f.Path, &f.Version,
			&f.CreatedAt, &f.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

func (q *Queries) UpdateFileVersion(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, updateFileVersion, id)
	return err
}

func (q *Queries) DeleteFile(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteFile, id)
	return err
}

func (q *Queries) DeleteFilesByChatID(ctx context.Context, chatID string) error {
	_, err := q.db.ExecContext(ctx, deleteFilesByChatID, chatID)
	return err
}

// Read file queries

func (q *Queries) RecordReadFile(ctx context.Context, arg RecordReadFileParams) error {
	_, err := q.db.ExecContext(ctx, recordReadFile, arg.ChatID, arg.Path)
	return err
}

type RecordReadFileParams struct {
	ChatID string
	Path  string
}

func (q *Queries) GetReadFilesByChatID(ctx context.Context, chatID string) ([]ReadFile, error) {
	rows, err := q.db.QueryContext(ctx, getReadFilesByChatID, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files []ReadFile
	for rows.Next() {
		var f ReadFile
		err := rows.Scan(&f.ChatID, &f.Path, &f.ReadAt)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

func (q *Queries) GetRecentlyReadFiles(ctx context.Context, arg GetRecentlyReadFilesParams) ([]GetRecentlyReadFilesRow, error) {
	rows, err := q.db.QueryContext(ctx, getRecentlyReadFiles, arg.ChatID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var files []GetRecentlyReadFilesRow
	for rows.Next() {
		var f GetRecentlyReadFilesRow
		err := rows.Scan(&f.Path, &f.ReadAt)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

type GetRecentlyReadFilesParams struct {
	ChatID string
	Limit  int
}

type GetRecentlyReadFilesRow struct {
	Path  string
	ReadAt string
}

func (q *Queries) DeleteReadFilesByChatID(ctx context.Context, chatID string) error {
	_, err := q.db.ExecContext(ctx, deleteReadFilesByChatID, chatID)
	return err
}