package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{
		db: db,
	}
}

func (r *ChatRepository) GetChat(ctx context.Context, chatID uuid.UUID) (*model.Chat, error) {
	const query = `
		SELECT id, name
		FROM chats
		WHERE id = $1
	`

	var chat model.Chat
	err := r.db.QueryRowContext(ctx, query, chatID).Scan(
		&chat.ID,
		&chat.Name,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("chat not found")
	}
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

func (r *ChatRepository) GetUserChats(ctx context.Context, userID uuid.UUID) ([]*model.Chat, error) {
	const query = `
		SELECT c.id, c.name
		FROM users u
		LEFT JOIN chats c
		ON u.chat_id = c.id
		WHERE u.id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*model.Chat

	for rows.Next() {
		var chat model.Chat
		if err := rows.Scan(
			&chat.ID,
			&chat.Name,
		); err != nil {
			return nil, err
		}

		chats = append(chats, &chat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *ChatRepository) CreateChat(ctx context.Context, chat *model.Chat, creatorID uuid.UUID) error {
	const chatQuery = `
	INSERT INTO chats (id, name)
	VALUES ($1, $2)
	`
	const userQuery = `
	INSERT INTO users (id, chat_id, role)
	VALUES ($1, $2, $3)
	`
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, chatQuery, chat.ID, chat.Name); err != nil {
		return fmt.Errorf("create chat: %w", err)
	}
	if _, err = tx.ExecContext(ctx, userQuery, creatorID, chat.ID, model.Owner); err != nil {
		return fmt.Errorf("create owner: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *ChatRepository) UpdateChatName(ctx context.Context, chatID uuid.UUID, name string) error {
	const query = `
		UPDATE chats
		SET name = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, name, chatID)
	return err
}

func (r *ChatRepository) DeleteChat(ctx context.Context, chatID uuid.UUID) error {
	const query = `
		DELETE FROM chats
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, chatID)
	return err
}

func (r *ChatRepository) DoesChatExist(ctx context.Context, chatID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM chats
			WHERE id = $1
		)
	`

	var exists bool

	err := r.db.QueryRowContext(ctx, query, chatID).Scan(
		&exists,
	)

	if err != nil {
		return false, err
	}

	return exists, nil
}

var _ repository.ChatRepository = (*ChatRepository)(nil)
