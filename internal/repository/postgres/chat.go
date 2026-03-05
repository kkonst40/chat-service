package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	errs "github.com/kkonst40/ichat/internal/domain/errors"
	"github.com/kkonst40/ichat/internal/domain/model"
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
		SELECT id, name, is_group, last_message_at
		FROM chats
		WHERE id = $1
	`

	slog.DebugContext(ctx, "getting chat from DB", "chatID", chatID)

	var chat model.Chat
	err := r.db.QueryRowContext(ctx, query, chatID).Scan(
		&chat.ID,
		&chat.Name,
		&chat.IsGroup,
		&chat.LastMessageAt,
	)

	if err == sql.ErrNoRows {
		return nil, errs.ErrChatNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return &chat, nil
}

func (r *ChatRepository) GetUserChats(ctx context.Context, userID uuid.UUID, filter model.ChatFilter) ([]model.Chat, error) {
	const queryAll = `
		SELECT c.id, c.name, c.is_group, c.last_message_at
		FROM users u
		LEFT JOIN chats c
		ON u.chat_id = c.id
		WHERE u.id = $1
		ORDER BY c.last_message_at DESC
	`

	const queryOneOf = `
		SELECT c.id, c.name, c.is_group, c.last_message_at
		FROM users u
		LEFT JOIN chats c
		ON u.chat_id = c.id
		WHERE u.id = $1 AND c.is_group = $2
		ORDER BY c.last_message_at DESC
	`

	slog.DebugContext(ctx, "getting user chats from DB", "userID", userID)

	var rows *sql.Rows
	var err error

	switch filter {
	case model.AllChats:
		rows, err = r.db.QueryContext(ctx, queryAll, userID)
	case model.GroupChats:
		rows, err = r.db.QueryContext(ctx, queryOneOf, userID, true)
	case model.PersonalChats:
		rows, err = r.db.QueryContext(ctx, queryOneOf, userID, false)
	}

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
			&chat.IsGroup,
			&chat.LastMessageAt,
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
	const chatQuery = `
		INSERT INTO chats (id, name, is_group, last_message_at)
		VALUES ($1, $2, $3, $4)
	`
	const userQuery = `
		INSERT INTO users (id, chat_id, role)
		VALUES ($1, $2, $3)
	`

	slog.DebugContext(ctx, "creating new chat with creator user in DB", "chatID", chat.ID, "userID", creatorID)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(
		ctx,
		chatQuery,
		chat.ID,
		chat.Name,
		chat.IsGroup,
		chat.LastMessageAt,
	); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	if _, err = tx.ExecContext(
		ctx,
		userQuery,
		creatorID,
		chat.ID,
		model.Owner,
	); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return nil
}

func (r *ChatRepository) CreatePersonalChat(ctx context.Context, chat *model.Chat, userID1, userID2 uuid.UUID) error {
	const chatQuery = `
		INSERT INTO chats (id, name, is_group, last_message_at)
		VALUES ($1, $2, $3, $4)
	`
	const userQuery = `
		INSERT INTO users (id, chat_id, role)
		VALUES ($1, $2, $3) ($4, $5, $6)
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(
		ctx,
		chatQuery,
		chat.ID,
		chat.Name,
		chat.IsGroup,
		chat.LastMessageAt,
	); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	if _, err = tx.ExecContext(
		ctx,
		userQuery,
		userID1,
		chat.ID,
		model.Owner,
		userID2,
		chat.ID,
		model.Owner,
	); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return nil
}

func (r *ChatRepository) UpdateChatName(ctx context.Context, chatID uuid.UUID, name string) error {
	const query = `
		UPDATE chats
		SET name = $1
		WHERE id = $2
	`

	slog.DebugContext(ctx, "updating name of the chat in DB", "chatID", chatID, "new_name", name)

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
	const query = `
		DELETE FROM chats
		WHERE id = $1
	`

	slog.DebugContext(ctx, "deleting the chat from DB", "chatID", chatID)

	if _, err := r.db.ExecContext(ctx, query, chatID); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return nil
}

func (r *ChatRepository) DeletePersonalChat(ctx context.Context, userID1, userID2 uuid.UUID) error {
	const query = `
		DELETE FROM chats
		WHERE id = (
		    SELECT chat_id
		    FROM users
		    WHERE chat_id IN (
		        SELECT id FROM chats WHERE is_group = false
		    )
		    AND id IN ($1, $2)
		    GROUP BY chat_id
		    HAVING COUNT(DISTINCT id) = 2
		);
	`

	slog.DebugContext(ctx, "deleting personal chat from DB", "userID1", userID1, "userID2", userID2)

	if _, err := r.db.ExecContext(ctx, query, userID1, userID2); err != nil {
		return fmt.Errorf("%w: %w", errs.ErrDatabase, err)
	}

	return nil
}

func (r *ChatRepository) ChatExists(ctx context.Context, chatID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM chats
			WHERE id = $1
		)
	`

	slog.DebugContext(ctx, "checking if chat exists in DB", "chatID", chatID)

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
