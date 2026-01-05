package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/logger"
	"github.com/kkonst40/ichat/internal/model"
	"github.com/kkonst40/ichat/internal/repository"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

func (r *MessageRepository) GetMessage(ctx context.Context, msgID uuid.UUID) (*model.Message, error) {
	log := logger.FromContext(ctx)

	const query = `
		SELECT id, user_id, chat_id, text, created_at
		FROM messages
		WHERE id = $1
	`

	log.Debug("getting message from DB", "msgID", msgID)

	var msg model.Message
	err := r.db.QueryRowContext(ctx, query, msgID).Scan(
		&msg.ID,
		&msg.UserID,
		&msg.ChatID,
		&msg.Text,
		&msg.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message not found")
	}
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func (r *MessageRepository) GetChatMessages(ctx context.Context, chatID uuid.UUID) ([]*model.Message, error) {
	log := logger.FromContext(ctx)
	const query = `
		SELECT id, user_id, chat_id, text, created_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at ASC
	`

	log.Debug("getting chat messages from DB", "chatID", chatID)

	rows, err := r.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*model.Message

	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.UserID,
			&msg.ChatID,
			&msg.Text,
			&msg.CreatedAt,
		); err != nil {
			return nil, err
		}

		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepository) CreateMessage(ctx context.Context, msg *model.Message) error {
	log := logger.FromContext(ctx)
	const query = `
		INSERT INTO messages (id, user_id, chat_id, text, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	log.Debug("creating new message in DB", "msgID", msg.ID)

	_, err := r.db.ExecContext(
		ctx,
		query,
		msg.ID,
		msg.UserID,
		msg.ChatID,
		msg.Text,
		msg.CreatedAt,
	)

	return err
}

func (r *MessageRepository) UpdateMessage(ctx context.Context, msg *model.Message) error {
	log := logger.FromContext(ctx)
	const query = `
		UPDATE messages
		SET text = $1
		WHERE id = $2
	`

	log.Debug("updating message in DB", "msgID", msg.ID)

	_, err := r.db.ExecContext(ctx, query, msg.Text, msg.ID)
	return err
}

func (r *MessageRepository) DeleteMessage(ctx context.Context, msgID uuid.UUID) error {
	log := logger.FromContext(ctx)
	const query = `
		DELETE FROM messages
		WHERE id = $1
	`

	log.Debug("deleting message in DB", "msgID", msgID)

	_, err := r.db.ExecContext(ctx, query, msgID)
	return err
}

var _ repository.MessageRepository = (*MessageRepository)(nil)
