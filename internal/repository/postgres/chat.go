package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	errs "github.com/kkonst40/ichat/internal/domain/errors"
	"github.com/kkonst40/ichat/internal/domain/model"
	"github.com/kkonst40/ichat/internal/logger"
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
	log := logger.FromContext(ctx)
	const query = `
		SELECT id, name
		FROM chats
		WHERE id = $1
	`

	log.Debug("getting chat from DB", "chatID", chatID)

	var chat model.Chat
	err := r.db.QueryRowContext(ctx, query, chatID).Scan(
		&chat.ID,
		&chat.Name,
	)

	if err == sql.ErrNoRows {
		return nil, errs.ErrChatNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return &chat, nil
}

func (r *ChatRepository) GetUserChats(ctx context.Context, userID uuid.UUID) ([]model.Chat, error) {
	log := logger.FromContext(ctx)

	const query = `
		SELECT c.id, c.name
		FROM users u
		LEFT JOIN chats c
		ON u.chat_id = c.id
		WHERE u.id = $1
		ORDER BY c.last_message_at DESC
	`

	log.Debug("getting user chats from DB", "userID", userID)

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}
	defer rows.Close()

	var chats []model.Chat

	for rows.Next() {
		var chat model.Chat
		if err := rows.Scan(
			&chat.ID,
			&chat.Name,
		); err != nil {
			return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
		}

		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return chats, nil
}

func (r *ChatRepository) CreateChat(ctx context.Context, chat *model.Chat, creatorID uuid.UUID) error {
	log := logger.FromContext(ctx)
	const chatQuery = `
		INSERT INTO chats (id, name, last_message_at)
		VALUES ($1, $2, $3)
	`
	const userQuery = `
		INSERT INTO users (id, chat_id, role)
		VALUES ($1, $2, $3)
	`

	log.Debug("creating new chat with creator user in DB", "chatID", chat.ID, "userID", creatorID)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, chatQuery, chat.ID, chat.Name, time.Now()); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}
	if _, err = tx.ExecContext(ctx, userQuery, creatorID, chat.ID, model.Owner); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return nil
}

func (r *ChatRepository) UpdateChatName(ctx context.Context, chatID uuid.UUID, name string) error {
	log := logger.FromContext(ctx)
	const query = `
		UPDATE chats
		SET name = $1
		WHERE id = $2
	`

	log.Debug("updating name of the chat in DB", "chatID", chatID, "new_name", name)

	res, err := r.db.ExecContext(ctx, query, name, chatID)
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	if rowsAffected == 0 {
		return errs.ErrChatNotFound
	}

	return nil
}

func (r *ChatRepository) DeleteChat(ctx context.Context, chatID uuid.UUID) error {
	log := logger.FromContext(ctx)
	const query = `
		DELETE FROM chats
		WHERE id = $1
	`

	log.Debug("deleting the chat from DB", "chatID", chatID)

	if _, err := r.db.ExecContext(ctx, query, chatID); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return nil
}

func (r *ChatRepository) ChatExists(ctx context.Context, chatID uuid.UUID) (bool, error) {
	log := logger.FromContext(ctx)
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM chats
			WHERE id = $1
		)
	`

	log.Debug("checking if chat exists in DB", "chatID", chatID)

	var exists bool

	err := r.db.QueryRowContext(ctx, query, chatID).Scan(
		&exists,
	)

	if err != nil {
		return false, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return exists, nil
}

var _ repository.ChatRepository = (*ChatRepository)(nil)
