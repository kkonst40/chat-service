package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/apperror"
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
		return nil, &apperror.NotFoundError{Msg: fmt.Sprintf("message (%v) not found", msgID)}
	}
	if err != nil {
		return nil, &apperror.DBError{Msg: err.Error()}
	}

	return &msg, nil
}

func (r *MessageRepository) GetChatMessages(ctx context.Context, chatID uuid.UUID) ([]model.Message, error) {
	log := logger.FromContext(ctx)
	const query = `
		SELECT id, user_id, chat_id, text, created_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at DESC
	`

	log.Debug("getting chat messages from DB", "chatID", chatID)

	rows, err := r.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, &apperror.DBError{Msg: err.Error()}
	}
	defer rows.Close()

	var messages []model.Message

	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.UserID,
			&msg.ChatID,
			&msg.Text,
			&msg.CreatedAt,
		); err != nil {
			return nil, &apperror.DBError{Msg: err.Error()}
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, &apperror.DBError{Msg: err.Error()}
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

	if err != nil {
		return &apperror.DBError{Msg: err.Error()}
	}

	return nil
}

func (r *MessageRepository) UpdateMessage(ctx context.Context, msg *model.Message) error {
	log := logger.FromContext(ctx)
	const query = `
		UPDATE messages
		SET text = $1
		WHERE id = $2
	`

	log.Debug("updating message in DB", "msgID", msg.ID)

	res, err := r.db.ExecContext(ctx, query, msg.Text, msg.ID)
	if err != nil {
		return &apperror.DBError{Msg: err.Error()}
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return &apperror.DBError{Msg: err.Error()}
	}

	if rowsAffected == 0 {
		return &apperror.NotFoundError{Msg: fmt.Sprintf("message (%v) not found", msg.ID)}
	}

	return nil
}

func (r *MessageRepository) DeleteMessage(ctx context.Context, msgID uuid.UUID) error {
	log := logger.FromContext(ctx)
	const query = `
		DELETE FROM messages
		WHERE id = $1
	`

	log.Debug("deleting message in DB", "msgID", msgID)

	if _, err := r.db.ExecContext(ctx, query, msgID); err != nil {
		return &apperror.DBError{Msg: err.Error()}
	}

	return nil
}

var _ repository.MessageRepository = (*MessageRepository)(nil)
